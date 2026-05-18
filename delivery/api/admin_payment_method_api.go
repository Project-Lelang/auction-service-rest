package api

import (
	"net/http"
	"strconv"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type AdminPaymentMethodApi struct {
	*api
	paymentMethodUseCase use_case.PaymentMethodUseCase
}

func RegisterAdminPaymentMethodApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &AdminPaymentMethodApi{
		api:                  baseApi,
		paymentMethodUseCase: useCaseManager.PaymentMethodUseCase(),
	}

	adminPaymentGroup := router.Group("/admin/payment-methods")
	{
		adminPaymentGroup.POST("", api.CreatePaymentMethod())
		adminPaymentGroup.GET("", api.ListPaymentMethods())
		adminPaymentGroup.GET("/:id", api.GetPaymentMethodById())
		adminPaymentGroup.PATCH("/:id", api.UpdatePaymentMethod())
	}
}

// CreatePaymentMethod godoc
//
//	@Router		/admin/payment-methods [post]
//	@Summary	Admin — Create a new payment method
//	@tags		Admin Payment Methods
//	@Security	BearerAuth
//	@Param		body	body	dto_request.PaymentMethodCreateRequest	true	"Payment Method Data"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response
func (a *AdminPaymentMethodApi) CreatePaymentMethod() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.PaymentMethodCreateRequest
		ctx.mustBind(&request)

		data := a.paymentMethodUseCase.Create(ctx.context(), request)
		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"payment_method": data,
			},
		})
	})
}

// ListPaymentMethods godoc
//
//	@Router		/admin/payment-methods [get]
//	@Summary	Admin — Get list of payment methods
//	@tags		Admin Payment Methods
//	@Security	BearerAuth
//	@Param		name		query	string	false	"Filter by Name"
//	@Param		code		query	string	false	"Filter by Code"
//	@Param		type		query	string	false	"Filter by Type"
//	@Param		is_active	query	bool	false	"Filter by Activation Status"
//	@Param		page		query	int		false	"Page"
//	@Param		limit		query	int		false	"Limit"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminPaymentMethodApi) ListPaymentMethods() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.PaymentMethodFetchRequest
		ctx.mustBindQuery(&request)

		data, total := a.paymentMethodUseCase.Fetch(ctx.context(), request)
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

// GetPaymentMethodById godoc
//
//	@Router		/admin/payment-methods/{id} [get]
//	@Summary	Admin — Get detail of a payment method by ID
//	@tags		Admin Payment Methods
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Payment Method ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminPaymentMethodApi) GetPaymentMethodById() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		id, _ := strconv.ParseInt(ctx.getParam("id"), 10, 64)

		data := a.paymentMethodUseCase.GetById(ctx.context(), id)
		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"payment_method": data,
			},
		})
	})
}

// UpdatePaymentMethod godoc
//
//	@Router		/admin/payment-methods/{id} [patch]
//	@Summary	Admin — Update partial data / toggle status of a payment method
//	@tags		Admin Payment Methods
//	@Security	BearerAuth
//	@Param		id		path	int										true	"Payment Method ID"
//	@Param		body	body	dto_request.PaymentMethodUpdateRequest	true	"Partial Update Data"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response
func (a *AdminPaymentMethodApi) UpdatePaymentMethod() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		id, _ := strconv.ParseInt(ctx.getParam("id"), 10, 64)

		var request dto_request.PaymentMethodUpdateRequest
		ctx.mustBind(&request)

		data := a.paymentMethodUseCase.Update(ctx.context(), id, request)
		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"payment_method": data,
			},
		})
	})
}
