package config

import (
	"fmt"

	zk "github.com/rarimo/zkverifier-kit"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

const proofEventIDValue = "ac42d1a986804618c7a793fbe814d9b31e47be51e082806363dca6958f3062"

type VerifierConfig struct {
	VerificationKeys    map[string][]byte
	AllowedAge          int
	AllowedCitizenships []interface{} // more convenient to use for validation, replace on need
}

func (c *Config) Verifier() *zk.Verifier {
	return c.verifier.Do(func() interface{} {
		var cfg struct {
			VerificationKeyPath string   `fig:"verification_key_path,required"`
			AllowedAge          int      `fig:"allowed_age,required"`
			AllowedCitizenships []string `fig:"allowed_citizenships,required"`
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
			zk.WithRootVerifier(c.RootVerifier.RootVerifier()))

		if err != nil {
			panic(fmt.Errorf("failed to initialize passport verifier: %w", err))
		}

		return v
	}).(*zk.Verifier)
}
