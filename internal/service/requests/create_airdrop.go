package requests

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rarimo/airdrop-svc/resources"
)

func NewCreateAirdrop(r *http.Request) (req resources.CreateAirdropRequest, err error) {
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		err = newDecodeError("body", err)
		return
	}

	return req, validation.Errors{
		"data/id":                 validation.Validate(req.Data.ID, validation.Required),
		"data/type":               validation.Validate(req.Data.Type, validation.Required, validation.In(resources.CREATE_AIRDROP)),
		"data/attributes/address": validation.Validate(req.Data.Attributes.Address, validation.Required),
	}.Filter()
}

func newDecodeError(what string, err error) error {
	return validation.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}
