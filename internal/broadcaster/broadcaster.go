// Package broadcaster provides the functionality to broadcast transactions to
// the blockchain. It is similar to https://github.com/rarimo/broadcaster-svc,
// but is integrated into airdrop-svc purposely. The mentioned broadcaster does
// not allow you to track even successful transaction submission.
//
// The reason of broadcasting implementation is the same: account sequence
// (nonce) must be strictly incrementing in Cosmos.
package broadcaster

import (
	"context"
	"fmt"
	"time"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/data"
	ethermint "github.com/rarimo/rarimo-core/ethermint/types"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/distributed_lab/running"
)

const txCodeSuccess = 0

type Runner struct {
	log *logan.Entry
	q   *data.AirdropsQ
	config.Broadcaster
}

func Run(ctx context.Context, cfg *config.Config) {
	log := cfg.Log().WithField("service", "builtin-broadcaster")
	log.Info("Starting service")

	r := &Runner{
		log:         log,
		q:           data.NewAirdropsQ(cfg.DB().Clone()),
		Broadcaster: cfg.Broadcaster(),
	}

	running.WithBackOff(ctx, r.log, "builtin-broadcaster", r.run, 5*time.Second, 5*time.Second, 5*time.Second)
}

func (r *Runner) run(ctx context.Context) error {
	airdrops, err := r.q.New().FilterByStatus(data.TxStatusPending).Limit(r.QueryLimit).Select()
	if err != nil {
		return fmt.Errorf("select airdrops: %w", err)
	}
	if len(airdrops) == 0 {
		return nil
	}
	r.log.Debugf("Got %d pending airdrops, broadcasting now", len(airdrops))

	for _, drop := range airdrops {
		if err = r.handlePending(ctx, drop); err != nil {
			r.log.WithField("airdrop", drop).
				WithError(err).Error("Failed to handle pending airdrop")
			continue
		}
	}

	return nil
}

func (r *Runner) handlePending(ctx context.Context, airdrop data.Airdrop) (err error) {
	var txHash string

	defer func() {
		if err != nil {
			r.updateAirdropStatus(ctx, airdrop.ID, txHash, data.TxStatusFailed)
		}
	}()

	tx, err := r.createAirdropTx(ctx, airdrop)
	if err != nil {
		return fmt.Errorf("create airdrop tx: %w", err)
	}

	txHash, err = r.broadcastTx(ctx, tx)
	if err != nil {
		return fmt.Errorf("broadcast tx: %w", err)
	}

	r.updateAirdropStatus(ctx, airdrop.ID, txHash, data.TxStatusCompleted)
	return nil
}

func (r *Runner) createAirdropTx(ctx context.Context, airdrop data.Airdrop) ([]byte, error) {
	tx, err := r.genTx(ctx, 0, airdrop)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tx: %w", err)
	}

	gasUsed, err := r.simulateTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate tx: %w", err)
	}

	tx, err = r.genTx(ctx, gasUsed*3, airdrop)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tx after simulation: %w", err)
	}

	return tx, nil
}

func (r *Runner) genTx(ctx context.Context, gasLimit uint64, airdrop data.Airdrop) ([]byte, error) {
	tx, err := r.buildTransferTx(airdrop)
	if err != nil {
		return nil, fmt.Errorf("build transfer tx: %w", err)
	}

	builder, err := r.TxConfig.WrapTxBuilder(tx)
	if err != nil {
		return nil, fmt.Errorf("wrap tx with builder: %w", err)
	}
	builder.SetGasLimit(gasLimit)
	// there are no fees on the mainnet now, and applying fees requires a lot of work
	builder.SetFeeAmount(types.Coins{types.NewInt64Coin("urmo", 0)})

	resp, err := r.Auth.Account(ctx, &authtypes.QueryAccountRequest{Address: r.SenderAddress})
	if err != nil {
		return nil, fmt.Errorf("get sender account: %w", err)
	}

	var account ethermint.EthAccount
	if err = account.Unmarshal(resp.Account.Value); err != nil {
		return nil, fmt.Errorf("unmarshal sender account: %w", err)
	}

	err = builder.SetSignatures(signing.SignatureV2{
		PubKey: r.Sender.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  r.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: account.Sequence,
	})
	if err != nil {
		return nil, fmt.Errorf("set signatures to tx: %w", err)
	}

	signerData := xauthsigning.SignerData{
		ChainID:       r.ChainID,
		AccountNumber: account.AccountNumber,
		Sequence:      account.Sequence,
	}
	sigV2, err := clienttx.SignWithPrivKey(
		r.TxConfig.SignModeHandler().DefaultMode(), signerData,
		builder, r.Sender, r.TxConfig, account.Sequence,
	)
	if err != nil {
		return nil, fmt.Errorf("sign with private key: %w", err)
	}

	if err = builder.SetSignatures(sigV2); err != nil {
		return nil, fmt.Errorf("set signatures V2: %w", err)
	}

	return r.TxConfig.TxEncoder()(builder.GetTx())
}

func (r *Runner) simulateTx(ctx context.Context, tx []byte) (gasUsed uint64, err error) {
	sim, err := r.TxClient.Simulate(ctx, &client.SimulateRequest{TxBytes: tx})
	if err != nil {
		return 0, fmt.Errorf("simulate tx: %w", err)
	}

	r.log.Debugf("Gas wanted: %d; gas used in simulation: %d", sim.GasInfo.GasWanted, sim.GasInfo.GasUsed)
	return sim.GasInfo.GasUsed, nil
}

func (r *Runner) broadcastTx(ctx context.Context, tx []byte) (string, error) {
	grpcRes, err := r.TxClient.BroadcastTx(ctx, &client.BroadcastTxRequest{
		Mode:    client.BroadcastMode_BROADCAST_MODE_BLOCK,
		TxBytes: tx,
	})
	if err != nil {
		return "", fmt.Errorf("send tx: %w", err)
	}
	r.log.Debugf("Submitted transaction to the core: %s", grpcRes.TxResponse.TxHash)

	if grpcRes.TxResponse.Code != txCodeSuccess {
		return grpcRes.TxResponse.TxHash, fmt.Errorf("got error code: %d, info: %s, log: %s", grpcRes.TxResponse.Code, grpcRes.TxResponse.Info, grpcRes.TxResponse.RawLog)
	}

	return grpcRes.TxResponse.TxHash, nil
}

// If we don't update tx status from pending, having the successful funds
// transfer, it will be possible to double-spend. With this solution the
// double-spend may still occur, if the service is restarted before the
// successful update. There is a better solution with file creation on context
// cancellation and parsing it on start.
func (r *Runner) updateAirdropStatus(ctx context.Context, id, txHash, status string) {
	running.UntilSuccess(ctx, r.log, "tx-status-updater", func(_ context.Context) (bool, error) {
		var ptr *string
		if txHash != "" {
			ptr = &txHash
		}

		err := r.q.New().Update(id, map[string]any{
			"status":  status,
			"tx_hash": ptr,
		})

		return err == nil, err
	}, 2*time.Second, 10*time.Second)
}

func (r *Runner) buildTransferTx(airdrop data.Airdrop) (types.Tx, error) {
	tx := &bank.MsgSend{
		FromAddress: r.SenderAddress,
		ToAddress:   airdrop.Address,
		Amount:      r.AirdropCoins,
	}

	builder := r.TxConfig.NewTxBuilder()
	if err := builder.SetMsgs(tx); err != nil {
		return nil, fmt.Errorf("set messages: %w", err)
	}

	return builder.GetTx(), nil
}
