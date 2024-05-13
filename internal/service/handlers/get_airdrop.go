package handlers

import (
	"net/http"

	data "github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GetAirdrop(w http.ResponseWriter, r *http.Request) {
	nullifier, err := requests.NewGetAirdrop(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	airdrops, err := AirdropsQ(r).
		FilterByNullifier(nullifier).
		Select()
	if err != nil {
		Log(r).WithError(err).Error("Failed to select airdrops by nullifier")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if len(airdrops) == 0 {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	airdrop := airdrops[0]
	for _, a := range airdrops[1:] {
		if a.Status == data.TxStatusCompleted {
			airdrop = a
			break
		}
	}

	ape.Render(w, toAirdropResponse(airdrop))
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
