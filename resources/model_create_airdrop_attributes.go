/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type CreateAirdropAttributes struct {
	// Destination address for the airdrop
	Address string `json:"address"`
	// All passport-related data required to verify passport-based ZKP
	PassportData PassportData `json:"passport_data"`
}
