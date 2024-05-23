package config

import (
	"fmt"

	zk "github.com/rarimo/zkverifier-kit"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type GlobalParams struct {
	EventID       string
	QuerySelector string
	AirdropStart  int64
}

type Verifierer struct {
	Params     GlobalParams
	ZkVerifier *zk.Verifier
}

func (c *Config) Verifier() *Verifierer {
	return c.verifier.Do(func() interface{} {
		var cfg struct {
			VerificationKeyPath      string   `fig:"verification_key_path,required"`
			AllowedAge               int      `fig:"allowed_age,required"`
			AllowedCitizenships      []string `fig:"allowed_citizenships,required"`
			AllowedQuerySelector     string   `fig:"allowed_query_selector,required"`
			AllowedEventID           string   `fig:"allowed_event_id,required"`
			AllowedIdentityCount     int64    `fig:"allowed_identity_count,required"`
			AllowedIdentityTimestamp int64    `fig:"allowed_identity_timestamp,required"`
		}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "verifier")).
			Please()
		if err != nil {
			panic(fmt.Errorf("failed to figure out verifier: %w", err))
		}

		v, err := zk.NewPassportVerifier(nil,
			zk.WithVerificationKeyFile(cfg.VerificationKeyPath),
			zk.WithCitizenships(cfg.AllowedCitizenships...),
			zk.WithAgeAbove(cfg.AllowedAge),
			zk.WithProofSelectorValue(cfg.AllowedQuerySelector),
			zk.WithEventID(cfg.AllowedEventID),
			zk.WithIdentityVerifier(c.ProvideVerifier()),
			zk.WithIdentitiesCounter(cfg.AllowedIdentityCount),
			zk.WithIdentitiesCreationTimestampLimit(cfg.AllowedIdentityTimestamp),
		)

		if err != nil {
			panic(fmt.Errorf("failed to initialize passport verifier: %w", err))
		}

		return &Verifierer{
			ZkVerifier: v,
			Params: GlobalParams{
				AirdropStart:  cfg.AllowedIdentityTimestamp,
				EventID:       cfg.AllowedEventID,
				QuerySelector: cfg.AllowedQuerySelector,
			},
		}
	}).(*Verifierer)
}
