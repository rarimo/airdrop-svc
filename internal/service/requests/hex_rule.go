package requests

import (
	"encoding/hex"
	"fmt"
)

var isHex hexRule

type hexRule struct{}

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
