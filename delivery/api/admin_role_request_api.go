package api

import (
	"net/http"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type AdminRoleRequestApi struct {
	*api
	roleRequestUseCase use_case.RoleRequestUseCase
	userUseCase        use_case.UserUseCase
	userRoleUseCase    use_case.UserRoleUseCase
}

func RegisterAdminRoleRequestApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &AdminRoleRequestApi{
		api:                baseApi,
		roleRequestUseCase: useCaseManager.RoleRequestUseCase(),
		userUseCase:        useCaseManager.UserUseCase(),
		userRoleUseCase:    useCaseManager.UserRoleUseCase(),
	}

	router.POST("/admin/role-requests/list/:role", api.ListRoleRequests())

	userRequestGroup := router.Group("/admin/users/:userId")
	{
		// userRequestGroup.GET("/roles", api.GetUserRoles())
		userRequestGroup.GET("/role-requests", api.GetUserRoleRequests())
		userRequestGroup.PATCH("/role-requests/:requestId/approve", api.ApproveRoleRequest())
		userRequestGroup.PATCH("/role-requests/:requestId/reject", api.RejectRoleRequest())
	}
}

// ListRoleRequests godoc
//
//	@Router		/admin/role-requests/list/{role} [post]
//	@Summary	Admin — Get list of role requests by role
//	@tags		Admin Role Requests
//	@Security	BearerAuth
//	@Param		role	path	string	true	"Role (BIDDER/SELLER)"
//	@Accept		json
//	@Param		body	body	dto_request.RoleRequestFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminRoleRequestApi) ListRoleRequests() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {

		role := ctx.getParam("role")

		var request dto_request.RoleRequestFetchRequest

		ctx.mustBind(&request)

		request.Role = &role

		data, total := a.roleRequestUseCase.AdminFetch(
			ctx.context(),
			request,
		)
		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				data,
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// GetUserRoles godoc
//
//	@Router		/admin/users/{userId}/roles [get]
//	@Summary	Admin — Get active roles of a user
//	@tags		Admin Role Requests
//	@Security	BearerAuth
//	@Param		userId	path	int	true	"User ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
// func (a *AdminRoleRequestApi) GetUserRoles() gin.HandlerFunc {
// 	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
// 		userId := ctx.mustGetParamInt64("userId")

// 		// Panggil kembali userUseCase bawaan
// 		user := a.userUseCase.AdminGet(ctx.context(), userId)

// 		ctx.json(http.StatusOK, dto_response.Response{
// 			Data: dto_response.DataResponse{
// 				// Biasanya temanmu menggunakan field "Roles" (dengan R kapital) di struct-nya
// 				"roles": user.Roles,
// 			},
// 		})
// 	})
// }

// GetUserRoleRequests godoc
//
//	@Router		/admin/users/{userId}/role-requests [get]
//	@Summary	Admin — Get role request history of a user
//	@tags		Admin Role Requests
//	@Security	BearerAuth
//	@Param		userId	path	int	true	"User ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminRoleRequestApi) GetUserRoleRequests() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		userId := ctx.mustGetParamInt64("userId")

		data := a.roleRequestUseCase.AdminFetchByUserId(ctx.context(), userId)
		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"role_requests": data,
			},
		})
	})
}

// ApproveRoleRequest godoc
//
//	@Router		/admin/users/{userId}/role-requests/{requestId}/approve [patch]
//	@Summary	Admin — Approve a role request
//	@tags		Admin Role Requests
//	@Security	BearerAuth
//	@Param		userId		path	int	true	"User ID"
//	@Param		requestId	path	int	true	"Request ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.SuccessResponse
func (a *AdminRoleRequestApi) ApproveRoleRequest() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		requestId := ctx.mustGetParamInt64("requestId")

		a.roleRequestUseCase.Approve(ctx.context(), requestId)
		ctx.json(http.StatusOK, dto_response.SuccessResponse{Message: "Success"})
	})
}

// RejectRoleRequest godoc
//
//	@Router		/admin/users/{userId}/role-requests/{requestId}/reject [patch]
//	@Summary	Admin — Reject a role request
//	@tags		Admin Role Requests
//	@Security	BearerAuth
//	@Param		userId		path	int										true	"User ID"
//	@Param		requestId	path	int										true	"Request ID"
//	@Param		body		body	dto_request.RoleRequestRejectRequest	true	"Reject Reason"
//	@Produce	json
//	@Success	200	{object}	dto_response.SuccessResponse
func (a *AdminRoleRequestApi) RejectRoleRequest() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		requestId := ctx.mustGetParamInt64("requestId")

		var request dto_request.RoleRequestRejectRequest
		ctx.mustBind(&request)

		a.roleRequestUseCase.Reject(ctx.context(), requestId, request)
		ctx.json(http.StatusOK, dto_response.SuccessResponse{Message: "Success"})
	})
}
