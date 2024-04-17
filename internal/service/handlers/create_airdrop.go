package handlers

import (
	"fmt"
	"net/http"
	"strings"

	cosmos "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/iden3/go-rapidsnark/verifier"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

// Full list of the OpenSSL signature algorithms and hash-functions is provided here:
// https://www.openssl.org/docs/man1.1.1/man3/SSL_CTX_set1_sigalgs_list.html

const (
	SHA1   = "sha1"
	SHA256 = "sha256"

	SHA256withRSA   = "SHA256withRSA"
	SHA1withECDSA   = "SHA1withECDSA"
	SHA256withECDSA = "SHA256withECDSA"
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

	err = ParticipantsQ(r).Transaction(func() error {
		err = ParticipantsQ(r).Insert(nullifier, req.Data.Attributes.Address)
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

func verifyProof(req resources.CreateAirdropRequest, cfg *config.VerifierConfig) error {
	var key []byte
	algorithm := signatureAlgorithm(req.Data.Attributes.Algorithm)
	switch algorithm {
	case SHA1withECDSA:
		key = cfg.VerificationKeys[SHA1]
	case SHA256withRSA, SHA256withECDSA:
		key = cfg.VerificationKeys[SHA256]
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
		"ECDSA": SHA1withECDSA,
	},
	"SHA256": {
		"RSA":   SHA256withRSA,
		"ECDSA": SHA256withECDSA,
	},
}

func signatureAlgorithm(passedAlgorithm string) string {
	if passedAlgorithm == "rsaEncryption" {
		return SHA256withRSA
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
