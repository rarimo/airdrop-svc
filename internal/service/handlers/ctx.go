package handlers

import (
	"context"
	"net/http"

	"github.com/rarimo/airdrop-svc/internal/config"
	"github.com/rarimo/airdrop-svc/internal/data"
	zk "github.com/rarimo/zkverifier-kit"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	airdropsQCtxKey
	airdropAmountCtxKey
	verifierCtxKey
	airdropParamsCtxKey
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxAirdropsQ(q *data.AirdropsQ) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, airdropsQCtxKey, q)
	}
}

func AirdropsQ(r *http.Request) *data.AirdropsQ {
	return r.Context().Value(airdropsQCtxKey).(*data.AirdropsQ).New()
}

func CtxAirdropAmount(amount string) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, airdropAmountCtxKey, amount)
	}
}

func AirdropAmount(r *http.Request) string {
	return r.Context().Value(airdropAmountCtxKey).(string)
}

func CtxAirdropParams(amount config.GlobalParams) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, airdropParamsCtxKey, amount)
	}
}

func AirdropParams(r *http.Request) config.GlobalParams {
	return r.Context().Value(airdropParamsCtxKey).(config.GlobalParams)
}

func CtxVerifier(entry *zk.Verifier) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, verifierCtxKey, entry)
	}
}

func Verifier(r *http.Request) *zk.Verifier {
	return r.Context().Value(verifierCtxKey).(*zk.Verifier)
}
