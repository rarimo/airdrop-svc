package handlers

import (
	"context"
	"net/http"

	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	participantsQCtxKey
	verifierCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxParticipantsQ(q *data.ParticipantsQ) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, participantsQCtxKey, q)
	}
}

func ParticipantsQ(r *http.Request) *data.ParticipantsQ {
	return r.Context().Value(participantsQCtxKey).(*data.ParticipantsQ).New()
}

func CtxVerifier(entry *config.VerifierConfig) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, verifierCtxKey, entry)
	}
}

func Verifier(r *http.Request) *config.VerifierConfig {
	return r.Context().Value(verifierCtxKey).(*config.VerifierConfig)
}
