package api

import (
	"net/http"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/use_case"

	"github.com/gin-gonic/gin"
)

type AdminAuctionApi struct {
	*api
	auctionUseCase use_case.AuctionUseCase
}

// GetAdminDashboardReport godoc
//
//	@Router		/admin/auctions/dashboard-report [get]
//	@Summary	Get daily auction revenue and count report for admin dashboard
//	@tags		Admin Auction
//	@Security	BearerAuth
//	@Param		start_date	query	string	false	"Start date filter (YYYY-MM-DD)"
//	@Param		end_date	query	string	false	"End date filter (YYYY-MM-DD)"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=[]dto_response.DashboardDailyReport}
func (a *AdminAuctionApi) GetAdminDashboardReport() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleAdmin}, func(ctx apiContext) {
		var request dto_request.AdminDashboardReportRequest

		request.StartDate = ctx.getQuery("start_date")
		request.EndDate = ctx.getQuery("end_date")

		reports := a.auctionUseCase.GetAdminDashboardReport(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: reports,
		})
	})
}

func RegisterAdminAuctionApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	a := &AdminAuctionApi{
		api:            baseApi,
		auctionUseCase: useCaseManager.AuctionUseCase(),
	}

	// Group route khusus admin yang memerlukan autentikasi
	adminGroup := router.Group("/admin/auctions")
	{
		adminGroup.GET("/dashboard-report", a.GetAdminDashboardReport())
	}
}
