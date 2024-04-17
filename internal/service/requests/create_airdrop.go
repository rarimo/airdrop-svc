package requests

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
	val "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/resources"
)

const (
	PubSignalNullifier = iota
	pubSignalBirthDate
	pubSignalExpirationDate
	pubSignalCitizenship = 6
	pubSignalEventData   = 10
	pubSignalSelector    = 12

	proofSelectorValue = "placeholder" // todo configure
)

func NewCreateAirdrop(r *http.Request, cfg *config.VerifierConfig) (req resources.CreateAirdropRequest, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, newDecodeError("body", err)
	}

	var (
		attr          = req.Data.Attributes
		signals       = attr.ZkProof.PubSignals
		olderThanDate = time.Now().UTC().AddDate(-cfg.AllowedAge, 0, 0)
	)

	err = val.Errors{
		"data/type":                            val.Validate(req.Data.Type, val.Required, val.In(resources.CREATE_AIRDROP)),
		"data/attributes/address":              val.Validate(attr.Address, val.Required, isRarimoAddr),
		"data/attributes/algorithm":            val.Validate(attr.Algorithm, val.Required),
		"data/attributes/zk_proof/proof":       val.Validate(attr.ZkProof.Proof, val.Required),
		"data/attributes/zk_proof/pub_signals": val.Validate(signals, val.Required, val.Length(14, 14)),
	}.Filter()
	if err != nil {
		return req, err
	}

	addrHash, err := poseidonHash(attr.Address)
	if err != nil {
		return req, val.Errors{"data/attributes/address": err}
	}

	return req, val.Errors{
		"pub_signals/nullifier":       val.Validate(signals[PubSignalNullifier], val.Required),
		"pub_signals/selector":        val.Validate(signals[pubSignalSelector], val.Required, val.In(proofSelectorValue)),
		"pub_signals/expiration_date": val.Validate(signals[pubSignalExpirationDate], val.Required, afterDate(time.Now().UTC())),
		"pub_signals/birth_date":      val.Validate(signals[pubSignalBirthDate], val.Required, beforeDate(olderThanDate)),
		"pub_signals/citizenship":     val.Validate(signals[pubSignalCitizenship], val.Required, val.In(cfg.AllowedCitizenships...)),
		"pub_signals/event_data":      val.Validate(signals[pubSignalEventData], val.Required, val.In(addrHash, "0x"+addrHash)),
	}.Filter()
}

func poseidonHash(addr string) (string, error) {
	bytes, err := types.AccAddressFromBech32(addr)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}
	bigAddr := new(big.Int).SetBytes(bytes)

	h, err := poseidon.Hash([]*big.Int{bigAddr})
	if err != nil {
		return "", fmt.Errorf("hash address: %w", err)
	}

	return h.Text(16), nil
}

func newDecodeError(what string, err error) error {
	return val.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}
