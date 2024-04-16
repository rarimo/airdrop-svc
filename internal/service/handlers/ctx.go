package handlers

import (
	"context"
	"net/http"

	"github.com/rarimo/airdrop-svc/internal/config"
	data "github.com/rarimo/airdrop-svc/internal/data"
	"github.com/rarimo/saver-grpc-lib/broadcaster"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	participantsQCtxKey
	airdropAmountCtxKey
	broadcasterCtxKey
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

func CtxAirdropAmount(amount int64) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, airdropAmountCtxKey, amount)
	}
}

func AirdropAmount(r *http.Request) int64 {
	return r.Context().Value(airdropAmountCtxKey).(int64)
}

func CtxBroadcaster(broadcaster broadcaster.Broadcaster) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, broadcasterCtxKey, broadcaster)
	}
}

func Broadcaster(r *http.Request) broadcaster.Broadcaster {
	return r.Context().Value(broadcasterCtxKey).(broadcaster.Broadcaster)
}

func CtxVerifier(entry *config.VerifierConfig) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, verifierCtxKey, entry)
	}
}

func Verifier(r *http.Request) *config.VerifierConfig {
	return r.Context().Value(verifierCtxKey).(*config.VerifierConfig)
}
