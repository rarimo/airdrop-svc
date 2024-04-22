package service

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/go-chi/chi"
	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/service/handlers"
	"gitlab.com/distributed_lab/ape"
)

func Run(ctx context.Context, cfg *config.Config) {
	setBech32Prefixes()
	r := chi.NewRouter()

	r.Use(
		ape.RecoverMiddleware(cfg.Log()),
		ape.LoganMiddleware(cfg.Log()),
		ape.CtxMiddleware(
			handlers.CtxLog(cfg.Log()),
			handlers.CtxVerifier(cfg.Verifier()),
		),
		handlers.DBCloneMiddleware(cfg.DB()),
	)
	r.Route("/integrations/airdrop-svc/airdrops", func(r chi.Router) {
		r.Post("/", handlers.CreateAirdrop)
		r.Get("/{id}", handlers.GetAirdrop)
	})

	cfg.Log().Info("Service started")
	ape.Serve(ctx, r, cfg, ape.ServeOpts{})
}

func setBech32Prefixes() {
	c := types.GetConfig()
	c.SetBech32PrefixForAccount("rarimo", "rarimopub")
	c.SetBech32PrefixForValidator("rarimovaloper", "rarimovaloperpub")
	c.SetBech32PrefixForConsensusNode("rarimovalcons", "rarimovalconspub")
	c.Seal()
}
