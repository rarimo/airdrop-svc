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

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/rarimo/airdrop-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ethermint "github.com/rarimo/rarimo-core/ethermint/types"
	"gitlab.com/distributed_lab/running"
)

const txCodeSuccess = 0

type Runner struct {
	log          *logan.Entry
	participants *data.ParticipantsQ
	Config
}

// TODO: run in CLI
func Run(ctx context.Context, cfg config) {
	log := cfg.Log().WithField("service", "builtin-broadcaster")
	log.Info("Starting service")

	r := &Runner{
		log:          log,
		participants: data.NewParticipantsQ(cfg.DB().Clone()),
		Config:       cfg.Broadcaster(),
	}

	running.WithBackOff(ctx, r.log, "builtin-broadcaster", r.run, 5*time.Second, 5*time.Second, 5*time.Second)
}

func (r *Runner) run(ctx context.Context) error {
	participants, err := r.participants.New().Limit(r.queryLimit).Select()
	if err != nil {
		return fmt.Errorf("select participants: %w", err)
	}
	if len(participants) == 0 {
		return nil
	}
	r.log.Debugf("Got %d participants to broadcast airdrop transactions", len(participants))

	// TODO: handle errors: whether we should delete the participant or assign a failed status (hard)
	for _, participant := range participants {
		log := r.log.WithField("participant_nullifier", participant.Nullifier)

		tx, err := r.genTx(ctx, 0, participant)
		if err != nil {
			log.WithError(err).Error("Failed to generate tx")
			continue
		}

		gasUsed, err := r.simulateTx(ctx, tx)
		if err != nil {
			log.WithError(err).Error("Failed to simulate tx")
			continue
		}

		tx, err = r.genTx(ctx, gasUsed*3, participant)
		if err != nil {
			log.WithError(err).Error("Failed to generate tx after simulation")
			continue
		}

		if err = r.broadcastTx(ctx, tx); err != nil {
			log.WithError(err).Error("Failed to broadcast tx")
			continue
		}

		if err = r.participants.New().Delete(participant.Nullifier); err != nil {
			log.WithError(err).Error("Failed to delete successful tx")
			continue
		}
	}

	return nil
}

func (r *Runner) genTx(ctx context.Context, gasLimit uint64, p data.Participant) ([]byte, error) {
	tx, err := r.buildTransferTx(p)
	if err != nil {
		return nil, fmt.Errorf("build transfer tx: %w", err)
	}

	builder, err := r.txConfig.WrapTxBuilder(tx)
	if err != nil {
		return nil, fmt.Errorf("wrap tx with builder: %w", err)
	}
	builder.SetGasLimit(gasLimit)
	// there are no fees on the mainnet now, and applies fees requires a lot of work
	builder.SetFeeAmount(types.Coins{types.NewInt64Coin("urmo", 0)})

	resp, err := r.auth.Account(ctx, &authtypes.QueryAccountRequest{Address: r.senderAddress})
	if err != nil {
		return nil, fmt.Errorf("get sender account: %w", err)
	}

	var account ethermint.EthAccount
	if err = account.Unmarshal(resp.Account.Value); err != nil {
		return nil, fmt.Errorf("unmarshal sender account: %w", err)
	}

	err = builder.SetSignatures(signing.SignatureV2{
		PubKey: r.sender.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  r.txConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: account.Sequence,
	})
	if err != nil {
		return nil, fmt.Errorf("set signatures to tx: %w", err)
	}

	signerData := xauthsigning.SignerData{
		ChainID:       r.chainID,
		AccountNumber: account.AccountNumber,
		Sequence:      account.Sequence,
	}
	sigV2, err := clienttx.SignWithPrivKey(
		r.txConfig.SignModeHandler().DefaultMode(), signerData,
		builder, r.sender, r.txConfig, account.Sequence,
	)
	if err != nil {
		return nil, fmt.Errorf("sign with private key: %w", err)
	}

	if err = builder.SetSignatures(sigV2); err != nil {
		return nil, fmt.Errorf("set signatures V2: %w", err)
	}

	return r.txConfig.TxEncoder()(builder.GetTx())
}

func (r *Runner) simulateTx(ctx context.Context, tx []byte) (gasUsed uint64, err error) {
	sim, err := r.txClient.Simulate(ctx, &client.SimulateRequest{TxBytes: tx})
	if err != nil {
		return 0, fmt.Errorf("simulate tx: %w", err)
	}

	r.log.Debugf("Gas wanted: %d; gas used in simulation: %d", sim.GasInfo.GasWanted, sim.GasInfo.GasUsed)
	return sim.GasInfo.GasUsed, nil
}

func (r *Runner) broadcastTx(ctx context.Context, tx []byte) error {
	grpcRes, err := r.txClient.BroadcastTx(ctx, &client.BroadcastTxRequest{
		Mode:    client.BroadcastMode_BROADCAST_MODE_BLOCK,
		TxBytes: tx,
	})
	if err != nil {
		return fmt.Errorf("send tx: %w", err)
	}
	r.log.Debugf("Submitted transaction to the core: %s", grpcRes.TxResponse.TxHash)

	if grpcRes.TxResponse.Code != txCodeSuccess {
		return fmt.Errorf("got error code: %d, info: %s, log: %s", grpcRes.TxResponse.Code, grpcRes.TxResponse.Info, grpcRes.TxResponse.RawLog)
	}

	return nil
}

func (r *Runner) buildTransferTx(p data.Participant) (types.Tx, error) {
	tx := &bank.MsgSend{
		FromAddress: r.senderAddress,
		ToAddress:   p.Address,
		Amount:      r.airdropCoins,
	}

	builder := r.txConfig.NewTxBuilder()
	if err := builder.SetMsgs(tx); err != nil {
		return nil, fmt.Errorf("set messages: %w", err)
	}

	return builder.GetTx(), nil
}
