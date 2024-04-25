package requests

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/decred/dcrd/bech32"
	val "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/resources"
)

const (
	PubSignalNullifier = iota
	pubSignalBirthDate
	pubSignalExpirationDate
	pubSignalCitizenship = 6
	pubSignalEventID     = 9
	pubSignalEventData   = 10
	pubSignalSelector    = 12

	proofSelectorValue = "39"
	proofEventIDValue  = "ac42d1a986804618c7a793fbe814d9b31e47be51e082806363dca6958f3062"
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

	eventID, ok := new(big.Int).SetString(signals[pubSignalEventID], 10)
	if !ok {
		return req, newDecodeError(
			"pub_signals/event_id",
			fmt.Errorf("setting string %s", signals[pubSignalEventID]),
		)
	}

	_, addrBytes, err := bech32.Decode(attr.Address)
	if err != nil {
		return req, newDecodeError("data/attributes/address", err)
	}

	var (
		addrDec     = encodeInt(addrBytes)
		citizenship = decodeInt(signals[pubSignalCitizenship])
	)

	return req, val.Errors{
		"pub_signals/nullifier":       val.Validate(signals[PubSignalNullifier], val.Required),
		"pub_signals/selector":        val.Validate(signals[pubSignalSelector], val.Required, val.In(proofSelectorValue)),
		"pub_signals/expiration_date": val.Validate(signals[pubSignalExpirationDate], val.Required, afterDate(time.Now().UTC())),
		"pub_signals/birth_date":      val.Validate(signals[pubSignalBirthDate], val.Required, beforeDate(olderThanDate)),
		"pub_signals/citizenship":     val.Validate(citizenship, val.Required, val.In(cfg.AllowedCitizenships...)),
		"pub_signals/event_id":        val.Validate(eventID.Text(16), val.Required, val.In(proofEventIDValue)),
		"pub_signals/event_data":      val.Validate(signals[pubSignalEventData], val.Required, val.In(addrDec)),
	}.Filter()
}

func newDecodeError(what string, err error) error {
	return val.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}
