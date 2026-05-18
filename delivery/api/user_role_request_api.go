package api

import (
	"net/http"
	"strconv"

	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type UserRoleRequestApi struct {
	*api
	roleRequestUseCase use_case.RoleRequestUseCase
}

func RegisterUserRoleRequestApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &UserRoleRequestApi{
		api:                baseApi,
		roleRequestUseCase: useCaseManager.RoleRequestUseCase(),
	}

	// Route khusus User menggunakan .Authorize
	// router.POST("/role-requests", api.UserCreateRequest())
	router.PUT("/role-requests/:requestId", api.UserReRequest())
}

// UserCreateRequest godoc
//
//	@Router		/role-requests [post]
//	@Summary	User — Create a new role request
//	@tags		User Role Requests
//	@Accept		json
//	@Security	BearerAuth
//	@Param		body	body	dto_request.RoleRequestCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response
// func (a *UserRoleRequestApi) UserCreateRequest() gin.HandlerFunc {
// 	return a.Authorize(func(ctx apiContext) {
// 		var request dto_request.RoleRequestCreateRequest
// 		ctx.mustBind(&request)

// 		res := a.roleRequestUseCase.UserCreate(ctx.context(), request)
// 		ctx.json(http.StatusCreated, dto_response.Response{
// 			Data: dto_response.DataResponse{"role_request": res},
// 		})
// 	})
// }

// UserReRequest godoc
//
//	@Router		/role-requests/{requestId} [put]
//	@Summary	User — Re-submit a rejected role request
//	@tags		User Role Requests
//	@Accept		json
//	@Security	BearerAuth
//	@Param		requestId	path	int	true	"Request ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *UserRoleRequestApi) UserReRequest() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		requestId, _ := strconv.ParseInt(ctx.getParam("requestId"), 10, 64)

		res := a.roleRequestUseCase.UserReRequest(ctx.context(), requestId)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{"role_request": res},
		})
	})
}
