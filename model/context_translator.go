package model

import (
	"context"

	internalI18n "auction-service/internal/i18n"
)

type translatorCtxKeyType string

var translatorCtxKey = translatorCtxKeyType("translator")

func SetTranslatorCtx(ctx context.Context, translator *internalI18n.Localizer) context.Context {
	return context.WithValue(ctx, translatorCtxKey, translator)
}

func MustGetTranslatorCtx(ctx context.Context) *internalI18n.Localizer {
	v, ok := ctx.Value(translatorCtxKey).(*internalI18n.Localizer)
	if !ok {
		return nil
	}
	return v
}
