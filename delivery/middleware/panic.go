package middleware

import (
	"fmt"
	"runtime/debug"

	"auction-service/constant"
	"auction-service/delivery/dto_response"
	"auction-service/infrastructure"

	internalI18n "auction-service/internal/i18n"

	"github.com/gin-gonic/gin"
)

func PanicHandler(router gin.IRouter, loggerStack infrastructure.LoggerStack) {
	router.Use(func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var (
					debugStackResponse *dto_response.ErrorResponse
					logMessage         = ""
				)

				switch v := r.(type) {
				case dto_response.ErrorResponse:
					instantResponse := v

					localizer, _ := ctx.MustGet("i18n").(*internalI18n.Localizer)
					if localizer != nil {
						localization := localizer.MustTranslate(instantResponse.Message, instantResponse.Contents...)
						instantResponse.Message = localization
					}

					ctx.AbortWithStatusJSON(instantResponse.Code, instantResponse)
					return

				case error:
					localizer, _ := ctx.MustGet("i18n").(*internalI18n.Localizer)

					switch v {
					case constant.ErrNotAuthenticated:
						instantResponse := dto_response.NewUnauthorizedErrorResponse(constant.LanguageSystemUnauthorized)
						if localizer != nil {
							localization := localizer.MustTranslate(instantResponse.Message)
							instantResponse.Message = localization
						}
						ctx.AbortWithStatusJSON(instantResponse.Code, instantResponse)
						return

					case constant.ErrForbidden:
						instantResponse := dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden)
						if localizer != nil {
							localization := localizer.MustTranslate(instantResponse.Message)
							instantResponse.Message = localization
						}
						ctx.AbortWithStatusJSON(instantResponse.Code, instantResponse)
						return

					default:
						logMessage = fmt.Sprintf("Captured error: %s", v.Error())
					}

				default:
					logMessage = fmt.Sprintf("Unhandled err type %T, Content: %+v", v, v)
				}

				if len(logMessage) != 0 {
					logMessage += "\n"
				}
				logMessage += string(debug.Stack())
				loggerStack.WriteAll(logMessage)

				if debugStackResponse == nil {
					debugStackResponse = dto_response.NewInternalServerErrorResponseP()
				}
				localizer, _ := ctx.MustGet("i18n").(*internalI18n.Localizer)
				if localizer != nil {
					debugStackResponse.Message = localizer.MustTranslate(debugStackResponse.Message)
				}
				ctx.AbortWithStatusJSON(debugStackResponse.Code, debugStackResponse)
			}
		}()

		ctx.Next()
	})
}
