/*
 * GENERATED. Do not modify. Your changes might be overwritten!
 */

package resources

import "github.com/iden3/go-rapidsnark/types"

type CreateAirdropAttributes struct {
	// Destination address for the airdrop
	Address string `json:"address"`
	// ZK-proof of the passport data
	ZkProof types.ZKProof `json:"zk_proof"`
}
