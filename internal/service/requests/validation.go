package requests

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/types"
)

var isRarimoAddr addressRule

type (
	addressRule struct{}
	timeRule    struct {
		point    time.Time
		isBefore bool
	}
)

func (r addressRule) Validate(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid type: %T, expected string", data)
	}

	_, err := types.AccAddressFromBech32(str)
	if err != nil {
		return fmt.Errorf("invalid bech32 address: %w", err)
	}

	return nil
}

func (r timeRule) Validate(date interface{}) error {
	raw, ok := date.(string)
	if !ok {
		return fmt.Errorf("invalid type: %T, expected string", date)
	}

	parsed, err := time.Parse("060102", decodeInt(raw))
	if err != nil {
		return fmt.Errorf("invalid date string: %w", err)
	}

	if r.isBefore && parsed.After(r.point) {
		return errors.New("date is too late")
	}

	if !r.isBefore && parsed.Before(r.point) {
		return errors.New("date is too early")
	}

	return nil
}

func beforeDate(point time.Time) timeRule {
	return timeRule{
		point:    point,
		isBefore: true,
	}
}

func afterDate(point time.Time) timeRule {
	return timeRule{
		point:    point,
		isBefore: false,
	}
}

func encodeInt(b []byte) string {
	return new(big.Int).SetBytes(b).String()
}

func decodeInt(s string) string {
	b, _ := new(big.Int).SetString(s, 10)
	return string(b.Bytes())
}
