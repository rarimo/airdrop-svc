package handlers

import (
	"net/http"

	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
)

func GetAirdropParams(w http.ResponseWriter, r *http.Request) {
	ape.Render(w, resources.AirdropParamsResponse{
		Data: resources.AirdropParams{
			Key: resources.Key{
				Type: resources.AIRDROP,
			},
			Attributes: resources.AirdropParamsAttributes{
				EventId:       AirdropParams(r).EventID,
				StartedAt:     AirdropParams(r).AirdropStart,
				QuerySelector: AirdropParams(r).QuerySelector,
			},
		},
	})
}
