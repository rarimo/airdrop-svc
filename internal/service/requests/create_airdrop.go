package requests

import (
	"encoding/json"
	"fmt"
	"net/http"

	val "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/airdrop-svc/resources"
)

func NewCreateAirdrop(r *http.Request) (req resources.CreateAirdropRequest, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = newDecodeError("body", err)
		return
	}

	passport := req.Data.Attributes.PassportData
	sod := passport.DocumentSOD

	return req, val.Errors{
		"data/id":                 val.Validate(req.Data.ID, val.Required, isHex),
		"data/type":               val.Validate(req.Data.Type, val.Required, val.In(resources.CREATE_AIRDROP)),
		"data/attributes/address": val.Validate(req.Data.Attributes.Address, val.Required, isHex),

		"data/attributes/passport_data/zkproof/proof":       val.Validate(passport.ZKProof.Proof, val.Required),
		"data/attributes/passport_data/zkproof/pub_signals": val.Validate(passport.ZKProof.PubSignals, val.Required),

		"data/attributes/passport_data/document_sod/algorithm":            val.Validate(sod.Algorithm, val.Required),
		"data/attributes/passport_data/document_sod/signed_attributes":    val.Validate(sod.SignedAttributes, val.Required, isHex),
		"data/attributes/passport_data/document_sod/encapsulated_content": val.Validate(sod.EncapsulatedContent, val.Required, isHex),
		"data/attributes/passport_data/document_sod/signature":            val.Validate(sod.Signature, val.Required, isHex),
		"data/attributes/passport_data/document_sod/pem_file":             val.Validate(sod.PemFile, val.Required),
	}.Filter()
}

func newDecodeError(what string, err error) error {
	return val.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}
