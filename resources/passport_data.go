package resources

import (
	"encoding/asn1"

	"github.com/iden3/go-rapidsnark/types"
)

type PassportData struct {
	ZKProof     types.ZKProof `json:"zkproof"`
	DocumentSOD DocumentSOD   `json:"document_sod"`
}

type DocumentSOD struct {
	SignedAttributes    string `json:"signed_attributes"`
	Algorithm           string `json:"algorithm"`
	Signature           string `json:"signature"`
	PemFile             string `json:"pem_file"`
	EncapsulatedContent string `json:"encapsulated_content"`
}

type DigestAttribute struct {
	ID     asn1.ObjectIdentifier
	Digest []asn1.RawValue `asn1:"set"`
}

type EncapsulatedData struct {
	Version             int
	PrivateKeyAlgorithm asn1.RawValue
	PrivateKey          struct {
		El1 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El2 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El3 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El4 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El5 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El6 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El7 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
		El8 struct {
			Integer  int
			OctetStr asn1.RawValue
		}
	}
}
