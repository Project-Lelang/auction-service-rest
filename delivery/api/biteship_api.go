package api

import (
	"context"
	"net/http"

	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/infrastructure"
	"auction-service/use_case"
	"auction-service/util"

	"github.com/gin-gonic/gin"
)

type BiteshipApi struct {
	*api
	biteshipUseCase use_case.BiteshipUseCase
}

// SearchAreas godoc
//
//	@Router		/biteship/areas/filter [post]
//	@Summary	Search Biteship area IDs by keyword (city/district/postal code)
//	@tags		Biteship
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.BiteshipSearchAreasRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{areas=[]dto_response.BiteshipAreaResponse}}
func (a *BiteshipApi) SearchAreas() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.BiteshipSearchAreasRequest
		ctx.mustBind(&request)

		areas := a.biteshipUseCase.SearchAreas(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"areas": util.ConvertArray(ctx.context(), areas, func(_ context.Context, a infrastructure.BiteshipArea) dto_response.BiteshipAreaResponse {
					return dto_response.NewBiteshipAreaResponse(a)
				}),
			},
		})
	})
}

func RegisterBiteshipApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	a := &BiteshipApi{
		api:             baseApi,
		biteshipUseCase: useCaseManager.BiteshipUseCase(),
	}

	group := router.Group("/biteship")
	group.POST("/areas/filter", a.SearchAreas())
}
