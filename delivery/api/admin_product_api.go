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

type AdminProductApi struct {
	*api
	productUseCase use_case.ProductUseCase
}

// Approve godoc
//
//	@Router		/admin/users/{userId}/products/{productId}/approve [patch]
//	@Summary	Admin — Approve/Verify a product
//	@tags		Admin Products
//	@Security	BearerAuth
//	@Param		userId		path	string	true	"User ID"
//	@Param		productId	path	string	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.ProductResponse}
func (a *AdminProductApi) Approve() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		// Extract URL parameters manually using the existing ctx.getParam tool
		var request dto_request.AdminProductApproveRequest
		request.UserId = ctx.getParam("userId")
		request.ProductId = ctx.getParam("productId")

		product := a.productUseCase.AdminApprove(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewProductResponse(ctx.context(), product),
		})
	})
}

// Reject godoc
//
//	@Router		/admin/users/{userId}/products/{productId}/reject [patch]
//	@Summary	Admin — Reject a product with feedback message
//	@tags		Admin Products
//	@Security	BearerAuth
//	@Param		userId		path	string									true	"User ID"
//	@Param		productId	path	string									true	"Product ID"
//	@Param		body		body	dto_request.AdminProductRejectRequest	true	"Rejection details"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.ProductResponse}
func (a *AdminProductApi) Reject() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminProductRejectRequest
		// Extract URL parameters manually
		request.UserId = ctx.getParam("userId")
		request.ProductId = ctx.getParam("productId")

		// Bind the JSON body parameter ("message") safely
		ctx.mustBind(&request)

		product := a.productUseCase.AdminReject(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewProductResponse(ctx.context(), product),
		})
	})
}

// Fetch godoc
//
//	@Router		/admin/products/filter [post]
//	@Summary	Admin — get all products (paginated)
//	@tags		Admin Products
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.AdminProductFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.ProductResponse}}
func (a *AdminProductApi) Fetch() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminProductFetchRequest
		ctx.mustBind(&request)

		products, total := a.productUseCase.AdminFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), products, dto_response.NewProductResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// FetchStatusHistories godoc
//
//	@Router		/admin/products/{productId}/histories [post]
//	@Summary	Admin — get status history for any product
//	@tags		Admin Products
//	@Security	BearerAuth
//	@Param		productId	path	int	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{histories=[]dto_response.ProductStatusHistoryResponse}}
func (a *AdminProductApi) FetchStatusHistories() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminProductFetchStatusHistoriesRequest
		request.ProductId = ctx.getParam("productId")

		histories := a.productUseCase.AdminFetchStatusHistories(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"histories": util.ConvertArray(ctx.context(), histories, dto_response.NewProductStatusHistoryResponse),
			},
		})
	})
}

func RegisterAdminProductApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &AdminProductApi{
		api:            baseApi,
		productUseCase: useCaseManager.ProductUseCase(),
	}

	// Dynamic RESTful matching parameters matching specified requirements
	router.PATCH("/admin/users/:userId/products/:productId/approve", api.Approve())
	router.PATCH("/admin/users/:userId/products/:productId/reject", api.Reject())

	routerGroup := router.Group("/admin/products")
	routerGroup.POST("/filter", api.Fetch())
	routerGroup.POST("/:productId/histories", api.FetchStatusHistories())
}
