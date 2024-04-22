package handlers

import (
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/iden3/go-rapidsnark/verifier"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

// Full list of the OpenSSL signature algorithms and hash-functions is provided here:
// https://www.openssl.org/docs/man1.1.1/man3/SSL_CTX_set1_sigalgs_list.html

const (
	sha256rsa   = "SHA256withRSA"
	sha1ecdsa   = "SHA1withECDSA"
	sha256ecdsa = "SHA256withECDSA"
)

func CreateAirdrop(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewCreateAirdrop(r, Verifier(r))
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	nullifier := req.Data.Attributes.ZkProof.PubSignals[requests.PubSignalNullifier]

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

	if err = verifyProof(req, Verifier(r)); err != nil {
		Log(r).WithError(err).Info("Invalid proof")
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"data/attributes/zk_proof": err,
		})...)
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

func verifyProof(req resources.CreateAirdropRequest, cfg *config.VerifierConfig) error {
	var key []byte
	algorithm := signatureAlgorithm(req.Data.Attributes.Algorithm)
	switch algorithm {
	case sha1ecdsa:
		key = cfg.VerificationKeys["sha1"]
	case sha256rsa, sha256ecdsa:
		key = cfg.VerificationKeys["sha256"]
	default:
		return fmt.Errorf("unsupported algorithm: %s", req.Data.Attributes.Algorithm)
	}

	proof := req.Data.Attributes.ZkProof
	if err := verifier.VerifyGroth16(proof, key); err != nil {
		return fmt.Errorf("verify groth16: %w", err)
	}

	return nil
}

var algorithmsMap = map[string]map[string]string{
	"SHA1": {
		"ECDSA": sha1ecdsa,
	},
	"SHA256": {
		"RSA":   sha256rsa,
		"ECDSA": sha256ecdsa,
	},
}

func signatureAlgorithm(passedAlgorithm string) string {
	if passedAlgorithm == "rsaEncryption" {
		return sha256rsa
	}

	if strings.Contains(strings.ToUpper(passedAlgorithm), "PSS") {
		return "" // RSA-PSS is not currently supported
	}

	for hashFunc, signatureAlgorithms := range algorithmsMap {
		if !strings.Contains(strings.ToUpper(passedAlgorithm), hashFunc) {
			continue
		}

		for signatureAlgo, algorithmName := range signatureAlgorithms {
			if strings.Contains(strings.ToUpper(passedAlgorithm), signatureAlgo) {
				return algorithmName
			}
		}
	}

	return ""
}
