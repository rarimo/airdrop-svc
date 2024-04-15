package config

import (
	"github.com/rarimo/saver-grpc-lib/broadcaster"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config struct {
	comfig.Logger
	pgdb.Databaser
	comfig.Listenerer
	broadcaster.Broadcasterer

	airdrop comfig.Once
	getter  kv.Getter
}

func New(getter kv.Getter) Config {
	return Config{
		getter:        getter,
		Databaser:     pgdb.NewDatabaser(getter),
		Listenerer:    comfig.NewListenerer(getter),
		Logger:        comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Broadcasterer: broadcaster.New(getter),
	}
}
