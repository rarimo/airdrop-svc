package handlers

import (
	"errors"
	"net/http"

	"github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	zk "github.com/rarimo/zkverifier-kit"
	"github.com/rarimo/zkverifier-kit/identity"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

// Full list of the OpenSSL signature algorithms and hash-functions is provided here:
// https://www.openssl.org/docs/man1.1.1/man3/SSL_CTX_set1_sigalgs_list.html

func CreateAirdrop(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewCreateAirdrop(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	nullifier := req.Data.Attributes.ZkProof.PubSignals[zk.PubSignalNullifier]

	airdrop, err := AirdropsQ(r).
		FilterByNullifier(nullifier).
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

	err = Verifier(r).VerifyProof(req.Data.Attributes.ZkProof, zk.WithAddress(req.Data.Attributes.Address))
	if err != nil {
		if errors.Is(err, identity.ErrContractCall) {
			Log(r).WithError(err).Error("Failed to verify proof")
			ape.RenderErr(w, problems.InternalError())
			return
		}

		Log(r).WithError(err).Info("Invalid proof")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	airdrop, err = AirdropsQ(r).Insert(data.Airdrop{
		Nullifier: nullifier,
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
