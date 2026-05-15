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
		productId := ctx.getParam("productId")

		histories := a.productUseCase.AdminFetchStatusHistories(ctx.context(), productId)

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

	routerGroup := router.Group("/admin/products")
	routerGroup.POST("/filter", api.Fetch())
	routerGroup.POST("/:productId/histories", api.FetchStatusHistories())
}
