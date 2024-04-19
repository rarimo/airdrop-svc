/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "time"

type AirdropAttributes struct {
	// Destination address for the airdrop
	Address string `json:"address"`
	// RFC3339 UTC timestamp of the airdrop creation
	CreatedAt time.Time `json:"created_at"`
	// Status of the airdrop transaction
	Status string `json:"status"`
	// RFC3339 UTC timestamp of the airdrop successful tx
	UpdatedAt time.Time `json:"updated_at"`
}
