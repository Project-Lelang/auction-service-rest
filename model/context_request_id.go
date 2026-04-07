package model

import (
	"context"

	"auction-service/constant"
)

type requestIdCtxKeyType string

var requestIdCtxKey = requestIdCtxKeyType("request-id")

func SetRequestIdCtx(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIdCtxKey, requestId)
}

func GetRequestIdCtx(ctx context.Context) (string, error) {
	v, ok := ctx.Value(requestIdCtxKey).(string)
	if !ok {
		return "", constant.ErrRequestIdNotFound
	}
	return v, nil
}

func MustGetRequestIdCtx(ctx context.Context) string {
	v, err := GetRequestIdCtx(ctx)
	if err != nil {
		panic(err)
	}
	return v
}
