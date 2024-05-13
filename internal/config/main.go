package config

import (
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config struct {
	comfig.Logger
	pgdb.Databaser
	comfig.Listenerer
	Broadcasterer

	airdrop  comfig.Once
	verifier comfig.Once
	getter   kv.Getter
}

func New(getter kv.Getter) *Config {
	return &Config{
		getter:        getter,
		Databaser:     pgdb.NewDatabaser(getter),
		Listenerer:    comfig.NewListenerer(getter),
		Logger:        comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Broadcasterer: NewBroadcaster(getter),
	}
}
