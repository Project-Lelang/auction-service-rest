package model

import (
	"context"

	"auction-service/constant"
)

// UserClaims holds the claims extracted from a JWT token.
type UserClaims struct {
	UserId int64
	Phone  string
	Roles  []string
}

// HasRole returns true if the user has the given role or is a SUPERADMIN.
func (u *UserClaims) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == constant.RoleSuperAdmin || r == role {
			return true
		}
	}
	return false
}

type userCtxKeyType string

var userCtxKey = userCtxKeyType("user")

func SetUserCtx(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, userCtxKey, claims)
}

func GetUserCtx(ctx context.Context) (*UserClaims, error) {
	v, ok := ctx.Value(userCtxKey).(*UserClaims)
	if !ok || v == nil {
		return nil, constant.ErrNotAuthenticated
	}
	return v, nil
}

func MustGetUserCtx(ctx context.Context) *UserClaims {
	v, err := GetUserCtx(ctx)
	if err != nil {
		panic(constant.ErrNotAuthenticated)
	}
	return v
}
