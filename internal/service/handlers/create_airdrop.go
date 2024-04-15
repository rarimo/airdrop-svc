package handlers

import (
	"fmt"
	"net/http"

	cosmos "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func CreateAirdrop(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewCreateAirdrop(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	participant, err := ParticipantsQ(r).Get(req.Data.ID)
	if err != nil {
		Log(r).WithError(err).Error("Failed to get participant by ID")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if participant != nil {
		ape.RenderErr(w, problems.Conflict())
		return
	}

	err = ParticipantsQ(r).Transaction(func() error {
		err = ParticipantsQ(r).Insert(req.Data.ID, req.Data.Attributes.Address)
		if err != nil {
			return fmt.Errorf("insert participant: %w", err)
		}
		return broadcastWithdrawalTx(req, r)
	})

	if err != nil {
		Log(r).WithError(err).Error("Failed to save and perform airdrop")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func broadcastWithdrawalTx(req resources.CreateAirdropRequest, r *http.Request) error {
	urmo := AirdropAmount(r)
	tx := &bank.MsgSend{
		FromAddress: Broadcaster(r).Sender(),
		ToAddress:   req.Data.Attributes.Address,
		Amount:      cosmos.NewCoins(cosmos.NewInt64Coin("urmo", urmo)),
	}

	err := Broadcaster(r).BroadcastTx(r.Context(), tx)
	if err != nil {
		return fmt.Errorf("broadcast withdrawal tx: %w", err)
	}

	return nil
}
