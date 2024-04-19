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
		id        = chi.URLParam(r, "id")
		err error = validation.Errors{"{id}": validation.Validate(id, validation.Required)}
	)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	participant, err := ParticipantsQ(r).Get(id)
	if err != nil {
		Log(r).WithError(err).Error("Failed to get participant by ID")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if participant == nil {
		ape.RenderErr(w, problems.NotFound())
		return
	}

	ape.Render(w, toAirdropResponse(*participant))
}

func toAirdropResponse(p data.Participant) resources.AirdropResponse {
	return resources.AirdropResponse{
		Data: resources.Airdrop{
			Key: resources.Key{
				ID:   p.Nullifier,
				Type: resources.AIRDROP,
			},
			Attributes: resources.AirdropAttributes{
				Address:   p.Address,
				Status:    p.Status,
				CreatedAt: p.CreatedAt,
				UpdatedAt: p.UpdatedAt,
			},
		},
	}
}
