package middleware

import (
	"auction-service/constant"
	"auction-service/model"
	"auction-service/util"

	"github.com/gin-gonic/gin"
)

func RequestIdHandler(router gin.IRouter) {
	router.Use(func(ctx *gin.Context) {
		requestId := util.NewUuid()

		ctx.Request = ctx.Request.WithContext(model.SetRequestIdCtx(ctx.Request.Context(), requestId))
		ctx.Header(constant.HeaderRequestIdKey, requestId)

		ctx.Next()
	})
}
