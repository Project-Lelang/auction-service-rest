package api

import (
	"net/http"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"
	"auction-service/util"

	"github.com/gin-gonic/gin"
)

type AdminUserApi struct {
	*api
	userUseCase     use_case.UserUseCase
	userRoleUseCase use_case.UserRoleUseCase
}

// Create godoc
//
//	@Router		/admin/users [post]
//	@Summary	Create a new admin user
//	@tags		Admin Users
//	@Accept		json
//	@Security	BearerAuth
//	@Param		body	body	dto_request.AdminUserCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.SuccessResponse
func (a *AdminUserApi) Create() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminUserCreateRequest
		ctx.mustBind(&request)

		a.userUseCase.AdminCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.SuccessResponse{Message: "Success"})
	})
}

// FetchAdmin godoc
//
//	@Router		/admin/users/admins/filter [post]
//	@Summary	Get paginated list of admin users only
//	@tags		Admin Users
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.AdminUserFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.UserResponse}}
func (a *AdminUserApi) FetchAdmin() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminUserFetchRequest

		ctx.mustBind(&request)

		// FORCE role agar hanya mengambil data Admin
		adminRole := constant.RoleAdmin
		request.Role = &adminRole

		// Panggil use-case seperti biasa
		users, total := a.userUseCase.AdminFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), users, dto_response.NewUserResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// Fetch godoc
//
//	@Router		/admin/users/filter [post]
//	@Summary	Get paginated list of users
//	@tags		Admin Users
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.AdminUserFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.UserResponse}}
func (a *AdminUserApi) Fetch() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminUserFetchRequest
		ctx.mustBind(&request)

		users, total := a.userUseCase.AdminFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), users, dto_response.NewUserResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// Get godoc
//
//	@Router		/admin/users/{userId} [get]
//	@Summary	Get user by ID with roles
//	@tags		Admin Users
//	@Security	BearerAuth
//	@Param		userId	path	int	true	"User ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{user=dto_response.UserResponse}}
func (a *AdminUserApi) Get() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminUserGetRequest
		request.UserId = ctx.getParam("userId")

		user := a.userUseCase.AdminGet(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"user": dto_response.NewUserResponse(ctx.context(), user),
			},
		})
	})
}

// Delete godoc
//
//	@Router		/admin/users/{userId} [delete]
//	@Summary	Soft-delete a user
//	@tags		Admin Users
//	@Security	BearerAuth
//	@Param		userId	path	int	true	"User ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.SuccessResponse
func (a *AdminUserApi) Delete() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminUserDeleteRequest
		request.UserId = ctx.getParam("userId")

		a.userUseCase.AdminDelete(ctx.context(), request)
		ctx.json(http.StatusOK, dto_response.SuccessResponse{Message: "Success"})
	})
}

func RegisterAdminUserApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := AdminUserApi{
		api:             baseApi,
		userUseCase:     useCaseManager.UserUseCase(),
		userRoleUseCase: useCaseManager.UserRoleUseCase(),
	}

	routerGroup := router.Group("/admin/users")
	routerGroup.POST("", api.Create())
	routerGroup.POST("/filter", api.Fetch())
	routerGroup.POST("/admins/filter", api.FetchAdmin())
	routerGroup.GET("/:userId", api.Get())
	routerGroup.DELETE("/:userId", api.Delete())
}
