package handlers

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	cosmos "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/iden3/go-rapidsnark/types"
	"github.com/iden3/go-rapidsnark/verifier"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/service/requests"
	"github.com/rarimo/airdrop-svc/resources"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

// Full list of the OpenSSL signature algorithms and hash-functions is provided here:
// https://www.openssl.org/docs/man1.1.1/man3/SSL_CTX_set1_sigalgs_list.html

const (
	SHA1   = "sha1"
	SHA256 = "sha256"

	SHA256withRSA   = "SHA256withRSA"
	SHA1withECDSA   = "SHA1withECDSA"
	SHA256withECDSA = "SHA256withECDSA"
)

func CreateAirdrop(w http.ResponseWriter, r *http.Request) {
	req, err := requests.NewCreateAirdrop(r)
	if err != nil {
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	participant, err := ParticipantsQ(r).Get(req.Data.ID)
	if err != nil {
		Log(r).WithError(err).Error("Failed to get participant by ID")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if participant != nil {
		ape.RenderErr(w, problems.Conflict())
		return
	}

	if err = verifyPassportData(req.Data.Attributes.PassportData, Verifier(r)); err != nil {
		Log(r).WithError(err).Info("Invalid passport data")
		ape.RenderErr(w, problems.BadRequest(validation.Errors{
			"data/attributes/passport_data": err,
		})...)
		return
	}

	err = ParticipantsQ(r).Transaction(func() error {
		err = ParticipantsQ(r).Insert(req.Data.ID, req.Data.Attributes.Address)
		if err != nil {
			return fmt.Errorf("insert participant: %w", err)
		}
		return broadcastWithdrawalTx(req, r)
	})

	if err != nil {
		Log(r).WithError(err).Error("Failed to save and perform airdrop")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func verifyPassportData(passport resources.PassportData, cfg *config.VerifierConfig) error {
	algorithm := signatureAlgorithm(passport.DocumentSOD.Algorithm)
	if algorithm == "" {
		return fmt.Errorf("invalid signature algorithm: %s", passport.DocumentSOD.Algorithm)
	}

	signedAttributes, _ := hex.DecodeString(passport.DocumentSOD.SignedAttributes)
	encapsulatedContent, _ := hex.DecodeString(passport.DocumentSOD.EncapsulatedContent)

	if err := validateSignedAttributes(passport.DocumentSOD, algorithm); err != nil {
		return fmt.Errorf("invalid signed attributes: %w", err)
	}

	cert, err := parseCertificate([]byte(passport.DocumentSOD.PemFile))
	if err != nil {
		return fmt.Errorf("parse certificate: %w", err)
	}

	if err = verifySignature(passport, cert, signedAttributes, algorithm); err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	var key []byte
	switch algorithm {
	case SHA1withECDSA:
		key = cfg.VerificationKeys[SHA1]
	case SHA256withRSA, SHA256withECDSA:
		key = cfg.VerificationKeys[SHA256]
	}

	if err = verifier.VerifyGroth16(passport.ZKProof, key); err != nil {
		return fmt.Errorf("verify groth16: %w", err)
	}

	var encapsulatedData resources.EncapsulatedData
	if _, err = asn1.Unmarshal(encapsulatedContent, &encapsulatedData); err != nil {
		return fmt.Errorf("unmarshal raw encapsulated content: %w", err)
	}

	if err = validatePubSignals(cfg, passport.ZKProof, encapsulatedData.PrivateKey.El1.OctetStr.Bytes); err != nil {
		return fmt.Errorf("invalid pub signals: %w", err)
	}

	if err = validateCert(cert, cfg.MasterCerts); err != nil {
		return fmt.Errorf("invalid certificate: %w", err)
	}

	return nil
}

var algorithmsMap = map[string]map[string]string{
	"SHA1": {
		"ECDSA": SHA1withECDSA,
	},
	"SHA256": {
		"RSA":   SHA256withRSA,
		"ECDSA": SHA256withECDSA,
	},
}

func signatureAlgorithm(passedAlgorithm string) string {
	if passedAlgorithm == "rsaEncryption" {
		return SHA256withRSA
	}

	if strings.Contains(strings.ToUpper(passedAlgorithm), "PSS") {
		return "" // RSA-PSS is not currently supported
	}

	for hashFunc, signatureAlgorithms := range algorithmsMap {
		if !strings.Contains(strings.ToUpper(passedAlgorithm), hashFunc) {
			continue
		}

		for signatureAlgo, algorithmName := range signatureAlgorithms {
			if strings.Contains(strings.ToUpper(passedAlgorithm), signatureAlgo) {
				return algorithmName
			}
		}
	}

	return ""
}

func validateSignedAttributes(sod resources.DocumentSOD, algorithm string) error {
	signedAttributes, _ := hex.DecodeString(sod.SignedAttributes)
	signedAttributesASN1 := make([]asn1.RawValue, 0)
	if _, err := asn1.UnmarshalWithParams(signedAttributes, &signedAttributesASN1, "set"); err != nil {
		return fmt.Errorf("unmarshal signed attributes to ASN1 with params: %w", err)
	}
	if len(signedAttributesASN1) == 0 {
		return errors.New("signed attributes count is 0")
	}

	digestAttr := resources.DigestAttribute{}
	if _, err := asn1.Unmarshal(signedAttributesASN1[len(signedAttributesASN1)-1].FullBytes, &digestAttr); err != nil {
		return fmt.Errorf("unmarshal ASN1 signed attributes to digest attribute: %w", err)
	}
	if len(digestAttr.Digest) == 0 {
		return errors.New("signed attributes digest values count is 0")
	}

	encapsulatedContent, _ := hex.DecodeString(sod.EncapsulatedContent)
	d := messageDigest(encapsulatedContent, algorithm)

	if !bytes.Equal(digestAttr.Digest[0].Bytes, d) {
		return errors.New("digest signed attribute is not equal to encapsulated content hash")
	}

	return nil
}

func parseCertificate(pemFile []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemFile)
	if block == nil {
		return nil, errors.New("invalid PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse X.509 certificate: %w", err)
	}

	return cert, nil
}

func verifySignature(req resources.PassportData, cert *x509.Certificate, signedAttributes []byte, algo string) error {
	signature, err := hex.DecodeString(req.DocumentSOD.Signature)
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	digest := messageDigest(signedAttributes, algo)

	switch algo {
	case SHA256withRSA:
		pubKey := cert.PublicKey.(*rsa.PublicKey)
		if err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest, signature); err != nil {
			return fmt.Errorf("invalid RSA signature: %w", err)
		}
	case SHA1withECDSA, SHA256withECDSA:
		pubKey := cert.PublicKey.(*ecdsa.PublicKey)
		if !ecdsa.VerifyASN1(pubKey, digest, signature) {
			return errors.New("invalid ECDSA signature")
		}
	}

	return nil
}

func messageDigest(data []byte, algo string) []byte {
	var h hash.Hash
	switch algo {
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	default:
		return nil
	}

	h.Write(data)
	return h.Sum(nil)
}

func validatePubSignals(cfg *config.VerifierConfig, zkp types.ZKProof, dg1 []byte) error {
	if err := validatePubSignalsDG1Hash(dg1, zkp.PubSignals); err != nil {
		return fmt.Errorf("invalid dg1 hash: %w", err)
	}
	if err := validatePubSignalsCurrentDate(zkp.PubSignals); err != nil {
		return fmt.Errorf("invalid current date: %w", err)
	}
	if err := validatePubSignalsAge(cfg, zkp.PubSignals[9]); err != nil {
		return fmt.Errorf("invalid age: %w", err)
	}
	return nil
}

func validatePubSignalsDG1Hash(dg1 []byte, pubSignals []string) error {
	ints, err := stringsToArrayBigInt([]string{pubSignals[0], pubSignals[1]})
	if err != nil {
		return fmt.Errorf("convert pub signals to big ints: %w", err)
	}

	hashBytes := make([]byte, 0)
	hashBytes = append(hashBytes, ints[0].Bytes()...)
	hashBytes = append(hashBytes, ints[1].Bytes()...)

	if !bytes.Equal(dg1, hashBytes) {
		return errors.New("encapsulated data and proof pub signals hashes are different")
	}

	return nil
}

func validatePubSignalsCurrentDate(pubSignals []string) error {
	year, err := strconv.Atoi(pubSignals[3])
	if err != nil {
		return fmt.Errorf("invalid year: %w", err)
	}

	month, err := strconv.Atoi(pubSignals[4])
	if err != nil {
		return fmt.Errorf("invalid month: %w", err)
	}

	day, err := strconv.Atoi(pubSignals[5])
	if err != nil {
		return fmt.Errorf("invalid day: %w", err)
	}

	currentTime := time.Now().UTC()

	if currentTime.Year() != (2000 + year) {
		return fmt.Errorf("invalid year, expected %d, got %d", currentTime.Year(), 2000+year)
	}

	if currentTime.Month() != time.Month(month) {
		return fmt.Errorf("invalid month, expected %d, got %d", currentTime.Month(), month)
	}

	if currentTime.Day() != day {
		return fmt.Errorf("invalid day, expected %d, got %d", currentTime.Day(), day)
	}

	return nil
}

func validatePubSignalsAge(cfg *config.VerifierConfig, agePubSignal string) error {
	age, err := strconv.Atoi(agePubSignal)
	if err != nil {
		return fmt.Errorf("age is not int: %w", err)
	}
	if age < cfg.AllowedAge {
		return errors.New("invalid age")
	}
	return nil
}

func validateCert(cert *x509.Certificate, masterCertsPem []byte) error {
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(masterCertsPem)

	foundCerts, err := cert.Verify(x509.VerifyOptions{Roots: roots})
	if err != nil {
		return fmt.Errorf("invalid certificate: %w", err)
	}

	if len(foundCerts) == 0 {
		return fmt.Errorf("invalid certificate: not ")
	}

	return nil
}

func stringsToArrayBigInt(publicSignals []string) ([]*big.Int, error) {
	p := make([]*big.Int, 0, len(publicSignals))
	for _, s := range publicSignals {
		sb, err := stringToBigInt(s)
		if err != nil {
			return nil, err
		}
		p = append(p, sb)
	}
	return p, nil
}

func stringToBigInt(s string) (*big.Int, error) {
	base := 10
	if bytes.HasPrefix([]byte(s), []byte("0x")) {
		base = 16
		s = strings.TrimPrefix(s, "0x")
	}
	n, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, fmt.Errorf("cannot parse string to *big.Int: %s (base=%d)", s, base)
	}
	return n, nil
}

func broadcastWithdrawalTx(req resources.CreateAirdropRequest, r *http.Request) error {
	urmo := AirdropAmount(r)
	tx := &bank.MsgSend{
		FromAddress: Broadcaster(r).Sender(),
		ToAddress:   req.Data.Attributes.Address,
		Amount:      cosmos.NewCoins(cosmos.NewInt64Coin("urmo", urmo)),
	}

	err := Broadcaster(r).BroadcastTx(r.Context(), tx)
	if err != nil {
		return fmt.Errorf("broadcast withdrawal tx: %w", err)
	}

	return nil
}
