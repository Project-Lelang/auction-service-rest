package api

import (
	"net/http"

	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type AuthApi struct {
	*api
	authUseCase use_case.AuthUseCase
}

// Login godoc
//
//	@Router		/auth/login [post]
//	@Summary	Login
//	@tags		Auth
//	@Accept		json
//	@Param		AuthLoginRequest	body	dto_request.AuthLoginRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.AuthTokenResponse}
func (a *AuthApi) Login() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthLoginRequest
		ctx.mustBind(&request)

		data := a.authUseCase.Login(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewAuthTokenResponse(data),
		})
	})
}

// Refresh godoc
//
//	@Router		/auth/refresh [post]
//	@Summary	Refresh Token
//	@tags		Auth
//	@Accept		json
//	@Param		AuthRefreshTokenRequest	body	dto_request.AuthRefreshTokenRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.AuthTokenResponse}
func (a *AuthApi) Refresh() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthRefreshTokenRequest
		ctx.mustBind(&request)

		data := a.authUseCase.Refresh(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewAuthTokenResponse(data),
		})
	})
}

// Logout godoc
//
//	@Router			/auth/logout [post]
//	@Summary		Logout
//	@tags			Auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto_response.SuccessResponse
func (a *AuthApi) Logout() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		a.authUseCase.Logout(ctx.context())

		ctx.json(http.StatusOK, dto_response.SuccessResponse{
			Message: "OK",
		})
	})
}

func RegisterAuthApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := AuthApi{
		api:         baseApi,
		authUseCase: useCaseManager.AuthUseCase(),
	}

	routerGroup := router.Group("/auth")
	routerGroup.POST("/login", api.Login())
	routerGroup.POST("/refresh", api.Refresh())
	routerGroup.POST("/logout", api.Logout())
}
