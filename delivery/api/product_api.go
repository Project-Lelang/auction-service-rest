package api

import (
	"net/http"

	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type ProductApi struct {
	*api
	productUseCase use_case.ProductUseCase
}

// Get godoc
//
//	@Router		/products/{productId} [get]
//	@Summary	Get a single product by ID (public)
//	@tags		Products
//	@Param		productId	path	int	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *ProductApi) Get() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.ProductGetRequest
		request.ProductId = ctx.getParam("productId")

		product := a.productUseCase.Get(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"product": dto_response.NewProductResponse(ctx.context(), product),
			},
		})
	})
}

func RegisterProductApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &ProductApi{
		api:            baseApi,
		productUseCase: useCaseManager.ProductUseCase(),
	}

	routerGroup := router.Group("/products")
	routerGroup.GET("/:productId", api.Get())
}
