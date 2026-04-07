package model

import (
	"context"
)

type userAccessTokenCtxKeyType string

var userAccessTokenCtxKey = userAccessTokenCtxKeyType("user-access-token")

func SetUserAccessTokenCtx(ctx context.Context, token *UserAccessToken) context.Context {
	return context.WithValue(ctx, userAccessTokenCtxKey, token)
}

func GetUserAccessTokenCtx(ctx context.Context) (*UserAccessToken, error) {
	v, ok := ctx.Value(userAccessTokenCtxKey).(*UserAccessToken)
	if !ok || v == nil {
		return nil, ErrDbtxNotFound
	}
	return v, nil
}
