/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "encoding/json"

type CreateAirdrop struct {
	Key
	Attributes CreateAirdropAttributes `json:"attributes"`
}
type CreateAirdropRequest struct {
	Data     CreateAirdrop `json:"data"`
	Included Included      `json:"included"`
}

type CreateAirdropListRequest struct {
	Data     []CreateAirdrop `json:"data"`
	Included Included        `json:"included"`
	Links    *Links          `json:"links"`
	Meta     json.RawMessage `json:"meta,omitempty"`
}

func (r *CreateAirdropListRequest) PutMeta(v interface{}) (err error) {
	r.Meta, err = json.Marshal(v)
	return err
}

func (r *CreateAirdropListRequest) GetMeta(out interface{}) error {
	return json.Unmarshal(r.Meta, out)
}

// MustCreateAirdrop - returns CreateAirdrop from include collection.
// if entry with specified key does not exist - returns nil
// if entry with specified key exists but type or ID mismatches - panics
func (c *Included) MustCreateAirdrop(key Key) *CreateAirdrop {
	var createAirdrop CreateAirdrop
	if c.tryFindEntry(key, &createAirdrop) {
		return &createAirdrop
	}
	return nil
}
