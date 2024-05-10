package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	data "github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetAirdrop(w http.ResponseWriter, r *http.Request) {
	var (
		nullifier = chi.URLParam(r, "nullifier")
		err       = validation.Errors{"{nullifier}": validation.Validate(nullifier, validation.Required)}.Filter()
	)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	airdrop, err := AirdropsQ(r).
		FilterByNullifier(nullifier).
		FilterByStatus(data.TxStatusCompleted).
		Get()
	if err != nil {
		Log(r).WithError(err).Error("Failed to get airdrop by ID")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if airdrop == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, toAirdropResponse(*airdrop))
}

func toAirdropResponse(tx data.Airdrop) resources.AirdropResponse {
	return resources.AirdropResponse{
		Data: resources.Airdrop{
			Key: resources.Key{
				ID:   tx.ID,
				Type: resources.AIRDROP,
			},
			Attributes: resources.AirdropAttributes{
				Nullifier: tx.Nullifier,
				Address:   tx.Address,
				TxHash:    tx.TxHash,
				Amount:    tx.Amount,
				Status:    tx.Status,
				CreatedAt: tx.CreatedAt,
				UpdatedAt: tx.UpdatedAt,
			},
		},
	}
}
