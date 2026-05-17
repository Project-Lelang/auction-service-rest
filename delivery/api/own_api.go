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

type OwnApi struct {
	*api
	productUseCase           use_case.ProductUseCase
	userUseCase              use_case.UserUseCase
	roleRequestUseCase       use_case.RoleRequestUseCase
	withdrawalRequestUseCase use_case.WithdrawalRequestUseCase
	auctionUseCase           use_case.AuctionUseCase
}

// Create godoc
//
//	@Router		/own/products [post]
//	@Summary	Create a new product listing
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.ProductCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Create() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.ProductCreateRequest
		ctx.mustBind(&request)

		product := a.productUseCase.OwnCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"product": dto_response.NewProductResponse(ctx.context(), product),
			},
		})
	})
}

// FetchProducts godoc
//
//	@Router		/own/products/filter [post]
//	@Summary	Get the authenticated user's own products (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnProductFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.ProductResponse}}
func (a *OwnApi) FetchProducts() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnProductFetchRequest
		ctx.mustBind(&request)

		products, total := a.productUseCase.OwnFetch(ctx.context(), request)

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
//	@Router		/own/products/{productId}/histories [post]
//	@Summary	Get status history for the caller's own product
//	@tags		Own
//	@Security	BearerAuth
//	@Param		productId	path	int	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{histories=[]dto_response.ProductStatusHistoryResponse}}
func (a *OwnApi) FetchStatusHistories() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnProductFetchStatusHistoriesRequest
		request.ProductId = ctx.getParam("productId")

		histories := a.productUseCase.FetchStatusHistories(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"histories": util.ConvertArray(ctx.context(), histories, dto_response.NewProductStatusHistoryResponse),
			},
		})
	})
}

// Get godoc
//
//	@Router		/own/products/{productId} [get]
//	@Summary	Get the authenticated user's own product by ID
//	@tags		Own
//	@Security	BearerAuth
//	@Param		productId	path	string	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Get() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnProductGetRequest
		request.ProductId = ctx.getParam("productId")

		product := a.productUseCase.OwnGet(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"product": dto_response.NewProductResponse(ctx.context(), product),
			},
		})
	})
}

// Update godoc
//
//	@Router		/own/products/{productId} [put]
//	@Summary	Update the authenticated user's own product (only allowed when DRAFT)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		productId	path	string								true	"Product ID"
//	@Param		body		body	dto_request.OwnProductUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Update() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnProductUpdateRequest
		ctx.mustBind(&request)
		request.ProductId = ctx.getParam("productId")

		product := a.productUseCase.OwnUpdate(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"product": dto_response.NewProductResponse(ctx.context(), product),
			},
		})
	})
}

// Request godoc
//
//	@Router		/own/products/{productId}/request [patch]
//	@Summary	Request the authenticated user's own product for review (DRAFT → REQUEST)
//	@tags		Own
//	@Security	BearerAuth
//	@Param		productId	path	string	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Request() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnProductRequestRequest
		request.ProductId = ctx.getParam("productId")

		product := a.productUseCase.OwnRequest(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"product": dto_response.NewProductResponse(ctx.context(), product),
			},
		})
	})
}

// GetProfile godoc
//
//	@Router		/own/profiles [get]
//	@Summary	Get the authenticated user's own profile
//	@tags		Own
//	@Security	BearerAuth
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{user=dto_response.UserResponse}}
func (a *OwnApi) GetProfile() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		user := a.userUseCase.OwnGet(ctx.context())

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"user": dto_response.NewUserResponse(ctx.context(), user),
			},
		})
	})
}

// UpdateProfile godoc
//
//	@Router		/own/profiles [put]
//	@Summary	Update the authenticated user's own profile
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnProfileUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{user=dto_response.UserResponse}}
func (a *OwnApi) UpdateProfile() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnProfileUpdateRequest
		ctx.mustBind(&request)

		user := a.userUseCase.OwnUpdate(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"user": dto_response.NewUserResponse(ctx.context(), user),
			},
		})
	})
}

// CreateRoleRequest godoc
//
//	@Router		/own/role-requests [post]
//	@Summary	Submit a new role request (BIDDER or SELLER)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnRoleRequestCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{role_request=dto_response.RoleRequestResponse}}
func (a *OwnApi) CreateRoleRequest() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.OwnRoleRequestCreateRequest
		ctx.mustBind(&request)

		roleRequest := a.roleRequestUseCase.OwnCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"role_request": dto_response.NewRoleRequestResponse(roleRequest),
			},
		})
	})
}

// CreateWithdrawalRequest godoc
//
//	@Router		/own/withdrawal-requests [post]
//	@Summary	Submit a new withdrawal request
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnWithdrawalRequestCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{withdrawal_request=dto_response.WithdrawalRequestResponse}}
func (a *OwnApi) CreateWithdrawalRequest() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.OwnWithdrawalRequestCreateRequest
		ctx.mustBind(&request)

		withdrawalRequest := a.withdrawalRequestUseCase.OwnCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"withdrawal_request": dto_response.NewWithdrawalRequestResponse(withdrawalRequest),
			},
		})
	})
}

// FetchAuctions godoc
//
//	@Router		/own/auctions/filter [post]
//	@Summary	Get the authenticated seller's own auctions (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnAuctionFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.AuctionResponse}}
func (a *OwnApi) FetchAuctions() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionFetchRequest
		ctx.mustBind(&request)

		auctions, total := a.auctionUseCase.OwnFetch(ctx.context(), request)

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
//	@Router		/own/auctions/{auctionId} [get]
//	@Summary	Get the authenticated seller's own auction by ID
//	@tags		Own
//	@Security	BearerAuth
//	@Param		auctionId	path	string	true	"Auction ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) GetAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionGetRequest
		request.AuctionId = ctx.getParam("auctionId")

		auction := a.auctionUseCase.OwnGet(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

// CreateAuction godoc
//
//	@Router		/own/auctions [post]
//	@Summary	Schedule a new auction for a verified product
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnAuctionCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) CreateAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionCreateRequest
		ctx.mustBind(&request)

		auction := a.auctionUseCase.OwnCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

// UpdateAuction godoc
//
//	@Router		/own/auctions/{auctionId} [put]
//	@Summary	Update a scheduled auction (only allowed when SCHEDULED)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		auctionId	path	string							true	"Auction ID"
//	@Param		body		body	dto_request.OwnAuctionUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) UpdateAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionUpdateRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.getParam("auctionId")

		auction := a.auctionUseCase.OwnUpdate(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

func RegisterOwnApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &OwnApi{
		api:                      baseApi,
		productUseCase:           useCaseManager.ProductUseCase(),
		userUseCase:              useCaseManager.UserUseCase(),
		roleRequestUseCase:       useCaseManager.RoleRequestUseCase(),
		withdrawalRequestUseCase: useCaseManager.WithdrawalRequestUseCase(),
		auctionUseCase:           useCaseManager.AuctionUseCase(),
	}

	routerGroup := router.Group("/own")

	routerProfileGroup := routerGroup.Group("/profiles")
	routerProfileGroup.GET("", api.GetProfile())
	routerProfileGroup.PUT("", api.UpdateProfile())

	routerProductGroup := routerGroup.Group("/products")
	routerProductGroup.POST("", api.Create())
	routerProductGroup.POST("/filter", api.FetchProducts())
	routerProductGroup.GET("/:productId", api.Get())
	routerProductGroup.PUT("/:productId", api.Update())
	routerProductGroup.PATCH("/:productId/request", api.Request())
	routerProductGroup.POST("/:productId/histories", api.FetchStatusHistories())

	routerGroup.POST("/role-requests", api.CreateRoleRequest())
	routerGroup.POST("/withdrawal-requests", api.CreateWithdrawalRequest())

	routerAuctionGroup := routerGroup.Group("/auctions")
	routerAuctionGroup.POST("/filter", api.FetchAuctions())
	routerAuctionGroup.GET("/:auctionId", api.GetAuction())
	routerAuctionGroup.POST("", api.CreateAuction())
	routerAuctionGroup.PUT("/:auctionId", api.UpdateAuction())
}
