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
	log          *logan.Entry
	participants *data.ParticipantsQ
	config.Broadcaster
}

func Run(ctx context.Context, cfg *config.Config) {
	log := cfg.Log().WithField("service", "builtin-broadcaster")
	log.Info("Starting service")

	r := &Runner{
		log:          log,
		participants: data.NewParticipantsQ(cfg.DB().Clone()),
		Broadcaster:  cfg.Broadcaster(),
	}

	running.WithBackOff(ctx, r.log, "builtin-broadcaster", r.run, 5*time.Second, 5*time.Second, 5*time.Second)
}

func (r *Runner) run(ctx context.Context) error {
	participants, err := r.participants.New().FilterByStatus(data.TxStatusPending).Limit(r.QueryLimit).Select()
	if err != nil {
		return fmt.Errorf("select participants: %w", err)
	}
	if len(participants) == 0 {
		return nil
	}
	r.log.Debugf("Got %d participants to broadcast airdrop transactions", len(participants))

	for _, participant := range participants {
		log := r.log.WithField("participant_nullifier", participant.Nullifier)

		if err := r.handleParticipant(ctx, participant); err != nil {
			log.WithError(err).Error("Failed to handle participant")
			continue
		}
	}

	return nil
}

func (r *Runner) handleParticipant(ctx context.Context, participant data.Participant) error {
	tx, err := r.createAirdropTx(ctx, participant)
	if err != nil {
		return fmt.Errorf("creating airdrop tx: %w", err)
	}

	txHash, err := r.broadcastTx(ctx, tx)
	if err != nil {
		err2 := r.participants.New().UpdateStatus(participant.Nullifier, txHash, data.TxStatusFailed)
		if err2 != nil {
			return fmt.Errorf("update participant failed tx status: %w (broadcast tx error: %w)", err2, err)
		}

		return fmt.Errorf("broadcast tx: %w", err)
	}

	err = r.participants.New().UpdateStatus(participant.Nullifier, txHash, data.TxStatusCompleted)
	if err != nil {
		return fmt.Errorf("update participant completed tx status: %w", err)
	}

	return nil
}

func (r *Runner) createAirdropTx(ctx context.Context, participant data.Participant) ([]byte, error) {
	tx, err := r.genTx(ctx, 0, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tx: %w", err)
	}

	gasUsed, err := r.simulateTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to simulate tx: %w", err)
	}

	tx, err = r.genTx(ctx, gasUsed*3, participant)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tx after simulation: %w", err)
	}

	return tx, nil
}

func (r *Runner) genTx(ctx context.Context, gasLimit uint64, p data.Participant) ([]byte, error) {
	tx, err := r.buildTransferTx(p)
	if err != nil {
		return nil, fmt.Errorf("build transfer tx: %w", err)
	}

	builder, err := r.TxConfig.WrapTxBuilder(tx)
	if err != nil {
		return nil, fmt.Errorf("wrap tx with builder: %w", err)
	}
	builder.SetGasLimit(gasLimit)
	// there are no fees on the mainnet now, and applies fees requires a lot of work
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

func (r *Runner) buildTransferTx(p data.Participant) (types.Tx, error) {
	tx := &bank.MsgSend{
		FromAddress: r.SenderAddress,
		ToAddress:   p.Address,
		Amount:      r.AirdropCoins,
	}

	builder := r.TxConfig.NewTxBuilder()
	if err := builder.SetMsgs(tx); err != nil {
		return nil, fmt.Errorf("set messages: %w", err)
	}

	return builder.GetTx(), nil
}
