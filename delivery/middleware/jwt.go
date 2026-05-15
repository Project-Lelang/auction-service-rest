package middleware

import (
	"auction-service/constant"
	"auction-service/model"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

func JWTHandler(router gin.IRouter, authUseCase use_case.AuthUseCase) {
	router.Use(func(ctx *gin.Context) {
		token := ctx.Request.Header.Get("Authorization")
		if token == "" {
			cookie, err := ctx.Request.Cookie("access_token")
			if err == nil {
				token = "Bearer " + cookie.Value
			}
		}

		if token == "" {
			ctx.Next()
			return
		}

		claims, err := authUseCase.Parse(token)
		if err != nil {
			if err != constant.ErrNotAuthenticated {
				panic(err)
			}
			ctx.Next()
			return
		}

		ctx.Request = ctx.Request.WithContext(model.SetUserCtx(ctx.Request.Context(), claims))
		ctx.Next()
	})
}
