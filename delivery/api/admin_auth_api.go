package api

import (
	"net/http"

	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type AdminAuthApi struct {
	*api
	authUseCase use_case.AuthUseCase
}

// Login godoc
//
//	@Router		/admin/auth/login [post]
//	@Summary	Admin Login
//	@tags		Admin Auth
//	@Accept		json
//	@Param		body	body	dto_request.AdminAuthLoginRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.AuthTokenResponse}
func (a *AdminAuthApi) Login() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AdminAuthLoginRequest
		ctx.mustBind(&request)

		token := a.authUseCase.AdminLogin(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewAuthTokenResponse(token),
		})
	})
}

func RegisterAdminAuthApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &AdminAuthApi{
		api:         baseApi,
		authUseCase: useCaseManager.AuthUseCase(),
	}

	routerGroup := router.Group("/admin/auth")
	routerGroup.POST("/login", api.Login())
}
