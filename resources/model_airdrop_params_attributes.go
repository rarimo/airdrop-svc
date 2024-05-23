/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

type AirdropParamsAttributes struct {
	// Event identifier that is generated during ZKP query creation
	EventId string `json:"event_id"`
	// Query selector that is used for proof generation
	QuerySelector string `json:"query_selector"`
	// Unix timestamp in seconds when airdrop event starts
	StartedAt int64 `json:"started_at"`
}
