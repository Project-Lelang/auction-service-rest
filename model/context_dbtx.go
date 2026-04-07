package model

import (
	"context"

	"auction-service/infrastructure"
)

type dbtxCtxKeyType string

var dbtxCtxKey = dbtxCtxKeyType("dbtx")

func SetDbtxCtx(ctx context.Context, dbtx infrastructure.DBTX) (context.Context, error) {
	if _, err := GetDbtxCtx(ctx); err == nil {
		return nil, ErrDbtxAlreadyExist
	}
	return context.WithValue(ctx, dbtxCtxKey, dbtx), nil
}

func GetDbtxCtx(ctx context.Context) (infrastructure.DBTX, error) {
	v, ok := ctx.Value(dbtxCtxKey).(infrastructure.DBTX)
	if !ok {
		return nil, ErrDbtxNotFound
	}
	return v, nil
}
