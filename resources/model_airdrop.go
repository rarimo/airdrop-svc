/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type Airdrop struct {
	Key
	Attributes AirdropAttributes `json:"attributes"`
}
type AirdropResponse struct {
	Data     Airdrop  `json:"data"`
	Included Included `json:"included"`
}

type AirdropListResponse struct {
	Data     []Airdrop       `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *AirdropListResponse) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *AirdropListResponse) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustAirdrop - returns Airdrop from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustAirdrop(key Key) *Airdrop {
	var airdrop Airdrop
	if c.tryFindEntry(key, &airdrop) {
		return &airdrop
	}
	return nil
}
