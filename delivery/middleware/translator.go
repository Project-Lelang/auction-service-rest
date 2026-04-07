package middleware

import (
	internalI18n "auction-service/internal/i18n"
	internalValidator "auction-service/internal/validator"
	"auction-service/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

func TranslatorHandler(router gin.IRouter) {
	router.Use(func(ctx *gin.Context) {
		var (
			matcher = language.NewMatcher([]language.Tag{
				language.English,
				language.Indonesian,
			})
			accept = ctx.GetHeader("Accept-Language")
		)

		tag, _ := language.MatchStrings(matcher, accept)
		locale := tag.String()

		// i18n translator
		translator := internalI18n.NewLocalizer(locale)
		ctx.Set("i18n", translator)
		ctx.Request = ctx.Request.WithContext(model.SetTranslatorCtx(ctx.Request.Context(), translator))

		// validator translator
		validatorTranslator := internalValidator.GetTranslator(locale)
		ctx.Request = ctx.Request.WithContext(model.SetValidatorTranslatorCtx(ctx.Request.Context(), validatorTranslator))

		ctx.Next()
	})
}
