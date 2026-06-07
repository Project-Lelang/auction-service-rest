package api

import (
	"net/http"

	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type UserAddressApi struct {
	*api
	userAddressUseCase use_case.UserAddressUseCase
}

// CreateUserAddress godoc
//
//	@Router		/user-addresses [post]
//	@Summary	Create a new user address
//	@tags		UserAddress
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.UserAddressCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{user_address=dto_response.UserAddressResponse}}
func (a *UserAddressApi) Create() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.UserAddressCreateRequest
		ctx.mustBind(&request)

		address := a.userAddressUseCase.OwnCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"user_address": dto_response.NewUserAddressResponse(ctx.context(), address),
			},
		})
	})
}

// GetUserAddress godoc
//
//	@Router		/user-addresses/{userAddressId} [get]
//	@Summary	Get a specific user address
//	@tags		UserAddress
//	@Security	BearerAuth
//	@Param		userAddressId	path	string	true	"User Address ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{user_address=dto_response.UserAddressResponse}}
func (a *UserAddressApi) Get() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		request := dto_request.UserAddressGetRequest{
			UserAddressId: ctx.getParam("userAddressId"),
		}

		address := a.userAddressUseCase.Get(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"user_address": dto_response.NewUserAddressResponse(ctx.context(), address),
			},
		})
	})
}

// UpdateUserAddress godoc
//
//	@Router		/user-addresses/{userAddressId} [put]
//	@Summary	Update a user address
//	@tags		UserAddress
//	@Security	BearerAuth
//	@Accept		json
//	@Param		userAddressId	path	string									true	"User Address ID"
//	@Param		body			body	dto_request.UserAddressUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{user_address=dto_response.UserAddressResponse}}
func (a *UserAddressApi) Update() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.UserAddressUpdateRequest
		ctx.mustBind(&request)
		request.UserAddressId = ctx.getParam("userAddressId")

		address := a.userAddressUseCase.OwnUpdate(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"user_address": dto_response.NewUserAddressResponse(ctx.context(), address),
			},
		})
	})
}

// DeleteUserAddress godoc
//
//	@Router		/user-addresses/{userAddressId} [delete]
//	@Summary	Delete a user address
//	@tags		UserAddress
//	@Security	BearerAuth
//	@Param		userAddressId	path	string	true	"User Address ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.SuccessResponse}
func (a *UserAddressApi) Delete() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		request := dto_request.UserAddressDeleteRequest{
			UserAddressId: ctx.getParam("userAddressId"),
		}

		a.userAddressUseCase.OwnDelete(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.SuccessResponse{Message: "deleted"},
		})
	})
}

func RegisterUserAddressApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	a := &UserAddressApi{
		api:                baseApi,
		userAddressUseCase: useCaseManager.UserAddressUseCase(),
	}

	addressGroup := router.Group("/user-addresses")
	// Keep only GET detail here; other CRUD actions moved to /own for owner-only operations
	addressGroup.GET("/:userAddressId", a.Get())
}
