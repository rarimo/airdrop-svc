package config

import (
	"os"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

type VerifierConfig struct {
	VerificationKeys map[string][]byte
	MasterCerts      []byte
	AllowedAge       int
}

func (c *Config) Verifier() *VerifierConfig {
	return c.verifier.Do(func() interface{} {
		var cfg struct {
			VerificationKeysPaths map[string]string `fig:"verification_keys_paths,required"`
			MasterCertsPath       string            `fig:"master_certs_path,required"`
			AllowedAge            int               `fig:"allowed_age,required"`
		}

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks).
			From(kv.MustGetStringMap(c.getter, "verifier")).
			Please()
		if err != nil {
			panic(err)
		}

		verificationKeys := make(map[string][]byte)
		for algo, path := range cfg.VerificationKeysPaths {
			verificationKey, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}

			verificationKeys[algo] = verificationKey
		}

		masterCerts, err := os.ReadFile(cfg.MasterCertsPath)
		if err != nil {
			panic(err)
		}

		return &VerifierConfig{
			VerificationKeys: verificationKeys,
			MasterCerts:      masterCerts,
			AllowedAge:       cfg.AllowedAge,
		}
	}).(*VerifierConfig)
}
