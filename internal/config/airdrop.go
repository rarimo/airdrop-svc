package config

import (
	"fmt"

	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/kv"
)

func (c *Config) AirdropAmount() int64 {
	return c.airdrop.Do(func() interface{} {
		var cfg struct {
			Amount int64 `fig:"amount,required"`
		}

		err := figure.Out(&cfg).
			From(kv.MustGetStringMap(c.getter, "airdrop")).
			Please()
		if err != nil {
			panic(fmt.Errorf("failed to figure out airdrop amount: %w", err))
		}

		return cfg.Amount
	}).(int64)
}
