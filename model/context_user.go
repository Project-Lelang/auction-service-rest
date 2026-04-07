package model

import (
	"context"

	"auction-service/constant"
)

type userCtxKeyType string

var userCtxKey = userCtxKeyType("user")

func SetUserCtx(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func GetUserCtx(ctx context.Context) (*User, error) {
	v, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil, constant.ErrNotAuthenticated
	}
	return v, nil
}

func MustGetUserCtx(ctx context.Context) User {
	v, err := GetUserCtx(ctx)
	if err != nil {
		panic(err)
	}

	if v == nil || v.Id == "" {
		panic(constant.ErrNotAuthenticated)
	}

	return *v
}
