package service

import (
	"context"

	"github.com/go-chi/chi"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/service/handlers"
	"gitlab.com/distributed_lab/ape"
)

func Run(ctx context.Context, cfg config.Config) {
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(cfg.Log()),
		ape.LoganMiddleware(cfg.Log()),
		ape.CtxMiddleware(
			handlers.CtxLog(cfg.Log()),
			handlers.CtxAirdropAmount(cfg.AirdropAmount()),
			handlers.CtxBroadcaster(cfg.Broadcaster()),
		),
		handlers.DBCloneMiddleware(cfg.DB()),
	)
	r.Route("/integrations/airdrop-svc", func(r chi.Router) {
		r.Post("/airdrops", handlers.CreateAirdrop)
	})

	cfg.Log().Info("Service started")
	ape.Serve(ctx, r, cfg, ape.ServeOpts{})
}
