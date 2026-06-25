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
//	@Summary	User Login
//	@tags		Auth
//	@Accept		json
//	@Param		body	body	dto_request.AuthLoginRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.AuthTokenResponse}
func (a *AuthApi) Login() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthLoginRequest
		ctx.mustBind(&request)

		token := a.authUseCase.Login(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewAuthTokenResponse(token),
		})
	})
}

// RequestOtp godoc
//
//	@Router		/auth/request-otp [post]
//	@Summary	Request OTP for email verification
//	@tags		Auth
//	@Accept		json
//	@Param		body	body	dto_request.AuthOtpRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.SuccessResponse
func (a *AuthApi) RequestOtp() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthOtpRequest
		ctx.mustBind(&request)

		a.authUseCase.CreateOtp(ctx.context(), request.Email)

		ctx.json(http.StatusOK, dto_response.SuccessResponse{Message: "Success"})
	})
}

// ForgotPassword godoc
//
//	@Router		/auth/forgot-password [post]
//	@Summary	Request password reset OTP by email
//	@tags		Auth
//	@Accept		json
//	@Param		body	body	dto_request.AuthForgotPasswordRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.SuccessResponse
func (a *AuthApi) ForgotPassword() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthForgotPasswordRequest
		ctx.mustBind(&request)

		a.authUseCase.RequestForgotPassword(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.SuccessResponse{Message: "Success"})
	})
}

// ResetPassword godoc
//
//	@Router		/auth/reset-password [post]
//	@Summary	Reset password using email OTP
//	@tags		Auth
//	@Accept		json
//	@Param		body	body	dto_request.AuthResetPasswordRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.SuccessResponse
func (a *AuthApi) ResetPassword() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthResetPasswordRequest
		ctx.mustBind(&request)

		a.authUseCase.ResetPassword(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.SuccessResponse{Message: "Success"})
	})
}

// Register godoc
//
//	@Router		/auth/register [post]
//	@Summary	Register a new user (requires OTP)
//	@tags		Auth
//	@Accept		json
//	@Param		body	body	dto_request.AuthRegisterRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.SuccessResponse
func (a *AuthApi) Register() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuthRegisterRequest
		ctx.mustBind(&request)

		a.authUseCase.Register(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.SuccessResponse{Message: "Success"})
	})
}

// Save Fcm Token godoc
//
//	@Router		/auth/save-fcm-token [post]
//	@Summary	Save FCM token for a user
//	@tags		Auth
//	@Accept		json
//	@Param		body	body	dto_request.AuthFcmTokenRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.SuccessResponse
func (a *AuthApi) SaveFcmToken() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuthFcmTokenRequest
		ctx.mustBind(&request)

		a.authUseCase.SaveFcmToken(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.SuccessResponse{Message: "Success"})
	})
}

func RegisterAuthApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := AuthApi{
		api:         baseApi,
		authUseCase: useCaseManager.AuthUseCase(),
	}

	routerGroup := router.Group("/auth")
	routerGroup.POST("/login", api.Login())
	routerGroup.POST("/request-otp", api.RequestOtp())
	routerGroup.POST("/forgot-password", api.ForgotPassword())
	routerGroup.POST("/reset-password", api.ResetPassword())
	routerGroup.POST("/register", api.Register())
	routerGroup.POST("/save-fcm-token", api.SaveFcmToken())
}
