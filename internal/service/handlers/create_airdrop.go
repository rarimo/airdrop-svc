package handlers

import (
	"net/http"

	"github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	zk "github.com/rarimo/zkverifier-kit"
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

	participant, err := ParticipantsQ(r).Get(nullifier)
	if err != nil {
		Log(r).WithError(err).Error("Failed to get participant by ID")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if participant != nil {
		ape.RenderErr(w, problems.Conflict())
		return
	}

	err = Verifier(r).VerifyProof(req.Data.Attributes.ZkProof, zk.WithAddress(req.Data.Attributes.Address))
	if err != nil {
		Log(r).WithError(err).Info("Invalid proof")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	participant, err = ParticipantsQ(r).Insert(data.Participant{
		Nullifier: nullifier,
		Address:   req.Data.Attributes.Address,
		Status:    data.TxStatusPending,
		Amount:    AirdropAmount(r),
	})
	if err != nil {
		Log(r).WithError(err).WithField("nullifier", nullifier).Errorf("Failed to insert participant")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, toAirdropResponse(*participant))
}
