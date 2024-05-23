/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type AirdropParams struct {
	Key
	Attributes AirdropParamsAttributes `json:"attributes"`
}
type AirdropParamsResponse struct {
	Data     AirdropParams `json:"data"`
	Included Included      `json:"included"`
}

type AirdropParamsListResponse struct {
	Data     []AirdropParams `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *AirdropParamsListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *AirdropParamsListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustAirdropParams - returns AirdropParams from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustAirdropParams(key Key) *AirdropParams {
	var airdropParams AirdropParams
	if c.tryFindEntry(key, &airdropParams) {
		return &airdropParams
	}
	return nil
}
