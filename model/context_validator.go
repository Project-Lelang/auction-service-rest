package model

import (
	"context"

	ut "github.com/go-playground/universal-translator"
)

type validatorTranslatorCtxKeyType string

var validatorTranslatorCtxKey = validatorTranslatorCtxKeyType("validator-translator")

func SetValidatorTranslatorCtx(ctx context.Context, trans ut.Translator) context.Context {
	return context.WithValue(ctx, validatorTranslatorCtxKey, trans)
}

func GetValidatorTranslatorCtx(ctx context.Context) (ut.Translator, bool) {
	v, ok := ctx.Value(validatorTranslatorCtxKey).(ut.Translator)
	return v, ok
}
