package config

import (
	"fmt"
	"os"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type VerifierConfig struct {
	VerificationKeys    map[string][]byte
	AllowedAge          int
	AllowedCitizenships []interface{} // more convenient to use for validation, replace on need
}

func (c *Config) Verifier() *VerifierConfig {
	return c.verifier.Do(func() interface{} {
		var cfg struct {
			VerificationKeysPaths map[string]string `fig:"verification_keys_paths,required"`
			AllowedAge            int               `fig:"allowed_age,required"`
			AllowedCitizenships   []string          `fig:"allowed_citizenships,required"`
		}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "verifier")).
			Please()
		if err != nil {
			panic(fmt.Errorf("failed to figure out verifier: %w", err))
		}

		verificationKeys := make(map[string][]byte)
		for algo, path := range cfg.VerificationKeysPaths {
			verificationKey, err := os.ReadFile(path)
			if err != nil {
				panic(fmt.Errorf("failed to read verification key file: %w", err))
			}
			verificationKeys[algo] = verificationKey
		}

		citizenships := make([]interface{}, len(cfg.AllowedCitizenships))
		for i, ctz := range cfg.AllowedCitizenships {
			citizenships[i] = ctz
		}

		return &VerifierConfig{
			VerificationKeys:    verificationKeys,
			AllowedAge:          cfg.AllowedAge,
			AllowedCitizenships: citizenships,
		}
	}).(*VerifierConfig)
}
