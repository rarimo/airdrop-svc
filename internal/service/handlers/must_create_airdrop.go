package handlers

import (
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

// Full list of the OpenSSL signature algorithms and hash-functions is provided here:
// https://www.openssl.org/docs/man1.1.1/man3/SSL_CTX_set1_sigalgs_list.html

func MustCreateAirdrop(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewCreateAirdrop(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	pk, err := crypto.GenerateKey()
	if err != nil {
		Log(r).WithError(err).Error("failed to gene")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	airdrop, err := AirdropsQ(r).
		FilterByAddress(req.Data.Attributes.Address).
		FilterByStatus(data.TxStatusCompleted).
		Get()
	if err != nil {
		Log(r).WithError(err).Error("Failed to get airdrop by nullifier")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if airdrop != nil {
		ape.RenderErr(w, problems.Conflict())
		return
	}

	airdrop, err = AirdropsQ(r).Insert(data.Airdrop{
		Nullifier: pk.D.String(),
		Address:   req.Data.Attributes.Address,
		Amount:    AirdropAmount(r),
		Status:    data.TxStatusPending,
	})
	if err != nil {
		Log(r).WithError(err).Errorf("Failed to insert airdrop")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, toAirdropResponse(*airdrop))
}
