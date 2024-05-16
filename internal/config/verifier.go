package config

import (
	"fmt"

	zk "github.com/rarimo/zkverifier-kit"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

const proofEventIDValue = "304358862882731539112827930982999386691702727710421481944329166126417129570"

func (c *Config) Verifier() *zk.Verifier {
	return c.verifier.Do(func() interface{} {
		var cfg struct {
			VerificationKeyPath      string   `fig:"verification_key_path,required"`
			AllowedAge               int      `fig:"allowed_age,required"`
			AllowedCitizenships      []string `fig:"allowed_citizenships,required"`
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
			zk.WithEventID(proofEventIDValue),
			zk.WithRootVerifier(c.ProvideVerifier()),
			zk.WithIdentitiesCounter(cfg.AllowedIdentityCount),
			zk.WithIdentitiesCreationTimestampLimit(cfg.AllowedIdentityTimestamp),
		)

		if err != nil {
			panic(fmt.Errorf("failed to initialize passport verifier: %w", err))
		}

		return v
	}).(*zk.Verifier)
}
