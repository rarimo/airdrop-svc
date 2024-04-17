package requests

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var isHex hexRule

type (
	hexRule  struct{}
	timeRule struct {
		point    time.Time
		isBefore bool
	}
)

func (r hexRule) Validate(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("invalid type: %T, expected string", data)
	}

	_, err := hex.DecodeString(str)
	if err != nil {
		return fmt.Errorf("invalid hex string: %w", err)
	}

	return nil
}

func (r timeRule) Validate(date interface{}) error {
	str, ok := date.(string)
	if !ok {
		return fmt.Errorf("invalid type: %T, expected string", date)
	}

	parsed, err := time.Parse("060102", str)
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
		isBefore: false,
	}
}

func afterDate(point time.Time) timeRule {
	return timeRule{
		point:    point,
		isBefore: false,
	}
}
