package config

import (
	"github.com/rarimo/zkverifier-kit/csca"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config struct {
	comfig.Logger
	pgdb.Databaser
	comfig.Listenerer
	csca.RootVerifier
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
		RootVerifier:  csca.NewRootVerifier(getter),
		Broadcasterer: NewBroadcaster(getter),
	}
}
