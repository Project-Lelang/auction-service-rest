// package api/admin_withdrawal_request_api.go

package api

import (
	"net/http"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type AdminWithdrawalRequestApi struct {
	*api
	withdrawalRequestUseCase use_case.WithdrawalRequestUseCase
}

func RegisterAdminWithdrawalRequestApi(
	router gin.IRouter,
	baseApi *api,
	useCaseManager use_case.UseCaseManager,
) {

	api := &AdminWithdrawalRequestApi{
		api:                      baseApi,
		withdrawalRequestUseCase: useCaseManager.WithdrawalRequestUseCase(),
	}

	router.POST(
		"/admin/withdrawal-requests/filter",
		api.Fetch(),
	)

	// GET user withdrawal history - TANPA PAGINATION
	router.GET(
		"/admin/users/:userId/withdrawal-requests",
		api.FetchAllByUserId(),
	)

	router.PATCH(
		"/admin/users/:userId/withdrawal-requests/:withdrawalRequestId/complete",
		api.CompleteWithdrawalRequest(),
	)
}

// Fetch godoc
//
//	@Router		/admin/withdrawal-requests/filter [post]
//	@Summary	Admin — Get withdrawal requests by status
//	@tags		Admin Withdrawal Requests
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.WithdrawalRequestFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminWithdrawalRequestApi) Fetch() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.WithdrawalRequestFetchRequest
		ctx.mustBind(&request)

		data, total := a.withdrawalRequestUseCase.AdminFetch(
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

// FetchAllByUserId godoc
//
//	@Router		/admin/users/{userId}/withdrawal-requests [get]
//	@Summary	Admin — Get all withdrawal history by user ID (no pagination)
//	@tags		Admin Withdrawal Requests
//	@Security	BearerAuth
//	@Param		userId	path	int	true	"User ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=[]model.WithdrawalRequest}
func (a *AdminWithdrawalRequestApi) FetchAllByUserId() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		userId := ctx.mustGetParamInt64("userId")

		requests, err := a.withdrawalRequestUseCase.AdminFetchAllByUserId(
			ctx.context(),
			userId,
		)
		if err != nil {
			panic(err)
		}

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"nodes": requests,
				"total": len(requests),
			},
		})
	})
}

// CompleteWithdrawalRequest godoc
//
//	@Router		/admin/users/{userId}/withdrawal-requests/{withdrawalRequestId}/complete [patch]
//	@Summary	Admin — Complete withdrawal request
//	@tags		Admin Withdrawal Requests
//	@Security	BearerAuth
//	@Param		userId				path	int	true	"User ID"
//	@Param		withdrawalRequestId	path	int	true	"Withdrawal Request ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminWithdrawalRequestApi) CompleteWithdrawalRequest() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		userId := ctx.mustGetParamInt64("userId")
		withdrawalRequestId := ctx.mustGetParamInt64("withdrawalRequestId")

		admin := model.MustGetUserCtx(ctx.context())

		res := a.withdrawalRequestUseCase.Complete(
			ctx.context(),
			admin.UserId,
			userId,
			withdrawalRequestId,
		)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"withdrawal_request": res,
			},
		})
	})
}
