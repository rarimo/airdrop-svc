package requests

import (
	"net/http"

	"github.com/go-chi/chi"
	val "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func NewGetAirdrop(r *http.Request) (nullifier string, err error) {
	nullifier = chi.URLParam(r, "nullifier")

	err = val.Errors{
		"{nullifier}": val.Validate(nullifier, val.Required, is.Digit),
	}.Filter()

	return
}
