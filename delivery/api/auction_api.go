package api

import (
	"net/http"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/infrastructure"
	"auction-service/use_case"
	"auction-service/util"

	"github.com/gin-gonic/gin"
)

type AuctionApi struct {
	*api
	auctionUseCase  use_case.AuctionUseCase
	bidUseCase      use_case.BidUseCase
	winnerUseCase   use_case.WinnerUseCase
	paymentUseCase  use_case.PaymentUseCase
	shipmentUseCase use_case.ShipmentUseCase
}

// FetchAuctions godoc
//
//	@Router		/auctions/filter [post]
//	@Summary	List auctions (paginated, public)
//	@tags		Auction
//	@Accept		json
//	@Param		body	body	dto_request.AuctionFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.AuctionResponse}}
func (a *AuctionApi) FetchAuctions() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuctionFetchRequest
		ctx.mustBind(&request)

		auctions, total := a.auctionUseCase.Fetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), auctions, dto_response.NewAuctionResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// FetchOnGoingAuctions godoc
//
//	@Router		/auctions/on-going/filter [post]
//	@Summary	List ongoing and scheduled auctions (paginated, public)
//	@tags		Auction
//	@Accept		json
//	@Param		body	body	dto_request.AuctionFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.AuctionResponse}}
func (a *AuctionApi) FetchOnGoingAuctions() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuctionFetchRequest
		ctx.mustBind(&request)

		auctions, total := a.auctionUseCase.FetchOnGoingAndScheduled(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), auctions, dto_response.NewAuctionResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// GetAuction godoc
//
//	@Router		/auctions/{auctionId} [get]
//	@Summary	Get a single auction by ID (public)
//	@tags		Auction
//	@Param		auctionId	path	int	true	"Auction ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *AuctionApi) GetAuction() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var request dto_request.AuctionGetRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		auction := a.auctionUseCase.Get(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

// PlaceBid godoc
//
//	@Router		/auctions/{auctionId}/bids [post]
//	@Summary	Place a bid on an active auction (BIDDER only)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.AuctionBidCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{bid=dto_response.AuctionBidResponse}}
func (a *AuctionApi) PlaceBid() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.AuctionBidCreateRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		bid := a.bidUseCase.PlaceBid(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"bid": dto_response.NewAuctionBidResponse(ctx.context(), bid),
			},
		})
	})
}

// FetchBids godoc
//
//	@Router		/auctions/{auctionId}/bids/filter [post]
//	@Summary	List bids for an auction (authenticated users)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int									true	"Auction ID"
//	@Param		body		body	dto_request.AuctionBidFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.AuctionBidResponse}}
func (a *AuctionApi) FetchBids() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuctionBidFetchRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		bids, total := a.bidUseCase.FetchAuctionBids(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), bids, dto_response.NewAuctionBidResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// PlaceBidNoLocking godoc
//
//	@Router		/auctions/{auctionId}/bids/no-locking [post]
//	@Summary	Place a bid on an active auction (BIDDER only)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int									true	"Auction ID"
//	@Param		body		body	dto_request.AuctionBidCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{bid=dto_response.AuctionBidResponse}}
func (a *AuctionApi) PlaceBidNoLocking() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.AuctionBidCreateRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		bid := a.bidUseCase.PlaceBidNoLocking(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"bid": dto_response.NewAuctionBidResponse(ctx.context(), bid),
			},
		})
	})
}

// FetchWinners godoc
//
//	@Router		/auctions/{auctionId}/winners/filter [post]
//	@Summary	List winners for an auction (authenticated)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int										true	"Auction ID"
//	@Param		body		body	dto_request.AuctionWinnerFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.AuctionWinnerResponse}}
func (a *AuctionApi) FetchWinners() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuctionWinnerFetchRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		winners, total := a.winnerUseCase.FetchByAuction(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), winners, dto_response.NewAuctionWinnerResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// GetWinner godoc
//
//	@Router		/auctions/{auctionId}/winners/{winnerId} [get]
//	@Summary	Get a single winner record (authenticated)
//	@tags		Auction
//	@Security	BearerAuth
//	@Param		auctionId	path	int	true	"Auction ID"
//	@Param		winnerId	path	int	true	"Winner ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{winner=dto_response.AuctionWinnerResponse}}
func (a *AuctionApi) GetWinner() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuctionWinnerGetRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.WinnerId = ctx.mustGetParamInt64("winnerId")

		winner := a.winnerUseCase.GetByAuction(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"winner": dto_response.NewAuctionWinnerResponse(ctx.context(), winner),
			},
		})
	})
}

// GetPayment godoc
//
//	@Router		/auctions/{auctionId}/payments/{paymentId} [get]
//	@Summary	Get a single payment record (authenticated)
//	@tags		Auction
//	@Security	BearerAuth
//	@Param		auctionId	path	int	true	"Auction ID"
//	@Param		paymentId	path	int	true	"Payment ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{payment=dto_response.PaymentResponse}}
func (a *AuctionApi) GetPayment() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuctionPaymentGetRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.PaymentId = ctx.mustGetParamInt64("paymentId")

		payment := a.paymentUseCase.GetByAuction(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"payment": dto_response.NewPaymentResponse(ctx.context(), payment),
			},
		})
	})
}

// FetchShipments godoc
//
//	@Router		/auctions/{auctionId}/shipments/filter [post]
//	@Summary	List shipments for an auction (authenticated)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int										true	"Auction ID"
//	@Param		body		body	dto_request.AuctionShipmentFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{shipments=[]dto_response.ShipmentResponse}}
func (a *AuctionApi) FetchShipments() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuctionShipmentFetchRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		shipments := a.shipmentUseCase.FetchByAuction(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"shipments": util.ConvertArray(ctx.context(), shipments, dto_response.NewShipmentResponse),
			},
		})
	})
}

// GetShipment godoc
//
//	@Router		/auctions/{auctionId}/shipments/{shipmentId} [get]
//	@Summary	Get a single shipment record (authenticated)
//	@tags		Auction
//	@Security	BearerAuth
//	@Param		auctionId	path	int	true	"Auction ID"
//	@Param		shipmentId	path	int	true	"Shipment ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{shipment=dto_response.ShipmentResponse}}
func (a *AuctionApi) GetShipment() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.AuctionShipmentGetRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.ShipmentId = ctx.mustGetParamInt64("shipmentId")

		shipment := a.shipmentUseCase.GetByAuction(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"shipment": dto_response.NewShipmentResponse(ctx.context(), shipment),
			},
		})
	})
}

// Ship godoc
//
//	@Router		/auctions/{auctionId}/shipments/{shipmentId}/ship [post]
//	@Summary	Mark a shipment as shipped with a tracking number (SELLER only)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int										true	"Auction ID"
//	@Param		shipmentId	path	int										true	"Shipment ID"
//	@Param		body		body	dto_request.AuctionShipmentShipRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{shipment=dto_response.ShipmentResponse}}
func (a *AuctionApi) Ship() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.AuctionShipmentShipRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.ShipmentId = ctx.mustGetParamInt64("shipmentId")

		shipment := a.shipmentUseCase.Ship(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"shipment": dto_response.NewShipmentResponse(ctx.context(), shipment),
			},
		})
	})
}

// UpdateBidderAddress godoc
//
//	@Router		/auctions/{auctionId}/shipments/{shipmentId}/bidder-address [patch]
//	@Summary	Update bidder shipping address (BIDDER only, before shipped)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int												true	"Auction ID"
//	@Param		shipmentId	path	int												true	"Shipment ID"
//	@Param		body		body	dto_request.AuctionShipmentUpdateAddressRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{shipment=dto_response.ShipmentResponse}}
func (a *AuctionApi) UpdateBidderAddress() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.AuctionShipmentUpdateAddressRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.ShipmentId = ctx.mustGetParamInt64("shipmentId")

		shipment := a.shipmentUseCase.UpdateBidderAddress(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"shipment": dto_response.NewShipmentResponse(ctx.context(), shipment),
			},
		})
	})
}

// UpdateSellerAddress godoc
//
//	@Router		/auctions/{auctionId}/shipments/{shipmentId}/seller-address [patch]
//	@Summary	Update seller sender address (SELLER only, before shipped)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int												true	"Auction ID"
//	@Param		shipmentId	path	int												true	"Shipment ID"
//	@Param		body		body	dto_request.AuctionShipmentUpdateAddressRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{shipment=dto_response.ShipmentResponse}}
func (a *AuctionApi) UpdateSellerAddress() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.AuctionShipmentUpdateAddressRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.ShipmentId = ctx.mustGetParamInt64("shipmentId")

		shipment := a.shipmentUseCase.UpdateSellerAddress(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"shipment": dto_response.NewShipmentResponse(ctx.context(), shipment),
			},
		})
	})
}

// GetTracking godoc
//
//	@Router		/auctions/{auctionId}/shipments/{shipmentId}/tracking [get]
//	@Summary	Get live tracking info via Komship (BIDDER/SELLER only)
//	@tags		Auction
//	@Security	BearerAuth
//	@Param		auctionId	path	int	true	"Auction ID"
//	@Param		shipmentId	path	int	true	"Shipment ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=object}
func (a *AuctionApi) GetTracking() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		request := dto_request.AuctionShipmentGetTrackingRequest{
			AuctionId:  ctx.mustGetParamInt64("auctionId"),
			ShipmentId: ctx.mustGetParamInt64("shipmentId"),
		}

		result := a.shipmentUseCase.GetTracking(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"tracking": result,
			},
		})
	})
}

//	@Router		/auctions/{auctionId}/shipments/{shipmentId}/receive [post]
//	@Summary	Mark a shipment as received with delivery proof (BIDDER only)
//	@tags		Auction
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	int											true	"Auction ID"
//	@Param		shipmentId	path	int											true	"Shipment ID"
//	@Param		body		body	dto_request.AuctionShipmentReceiveRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{shipment=dto_response.ShipmentResponse}}
func (a *AuctionApi) Receive() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.AuctionShipmentReceiveRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")
		request.ShipmentId = ctx.mustGetParamInt64("shipmentId")

		shipment := a.shipmentUseCase.Receive(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"shipment": dto_response.NewShipmentResponse(ctx.context(), shipment),
			},
		})
	})
}

// HandlePaymentNotification godoc
//
//	@Router		/payment-notifications [post]
//	@Summary	Webhook endpoint for Midtrans payment notifications
//	@tags		Auction
//	@Accept		json
//	@Param		body	body	infrastructure.MidtransNotification	true	"Midtrans notification payload"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.SuccessResponse}
func (a *AuctionApi) HandlePaymentNotification() gin.HandlerFunc {
	return a.Guest(func(ctx apiContext) {
		var notification infrastructure.MidtransNotification
		ctx.mustBind(&notification)

		a.paymentUseCase.HandleMidtransNotification(ctx.context(), notification)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.SuccessResponse{Message: "OK"},
		})
	})
}

func RegisterAuctionApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	a := &AuctionApi{
		api:             baseApi,
		auctionUseCase:  useCaseManager.AuctionUseCase(),
		bidUseCase:      useCaseManager.BidUseCase(),
		winnerUseCase:   useCaseManager.WinnerUseCase(),
		paymentUseCase:  useCaseManager.PaymentUseCase(),
		shipmentUseCase: useCaseManager.ShipmentUseCase(),
	}

	auctionsGroup := router.Group("/auctions")
	auctionsGroup.POST("/filter", a.FetchAuctions())
	auctionsGroup.POST("/on-going/filter", a.FetchOnGoingAuctions())
	auctionsGroup.GET("/:auctionId", a.GetAuction())
	auctionsGroup.POST("/:auctionId/bids/filter", a.FetchBids())
	auctionsGroup.POST("/:auctionId/bids", a.PlaceBid())
	auctionsGroup.POST("/:auctionId/bids/no-locking", a.PlaceBidNoLocking())
	auctionsGroup.POST("/:auctionId/winners/filter", a.FetchWinners())
	auctionsGroup.GET("/:auctionId/winners/:winnerId", a.GetWinner())
	auctionsGroup.GET("/:auctionId/payments/:paymentId", a.GetPayment())
	auctionsGroup.POST("/:auctionId/shipments/filter", a.FetchShipments())
	auctionsGroup.GET("/:auctionId/shipments/:shipmentId", a.GetShipment())
	auctionsGroup.POST("/:auctionId/shipments/:shipmentId/ship", a.Ship())
	auctionsGroup.POST("/:auctionId/shipments/:shipmentId/receive", a.Receive())
	auctionsGroup.PATCH("/:auctionId/shipments/:shipmentId/bidder-address", a.UpdateBidderAddress())
	auctionsGroup.PATCH("/:auctionId/shipments/:shipmentId/seller-address", a.UpdateSellerAddress())
	auctionsGroup.GET("/:auctionId/shipments/:shipmentId/tracking", a.GetTracking())

	router.POST("/payment-notifications", a.HandlePaymentNotification())
}
