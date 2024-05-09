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
		return req, newDecodeError("body", err)
	}

	attr := req.Data.Attributes
	return req, val.Errors{
		"data/type":               val.Validate(req.Data.Type, val.Required, val.In(resources.CREATE_AIRDROP)),
		"data/attributes/address": val.Validate(attr.Address, val.Required, isRarimoAddr),
	}.Filter()
}

func newDecodeError(what string, err error) error {
	return val.Errors{
		what: fmt.Errorf("decode request %s: %w", what, err),
	}
}
