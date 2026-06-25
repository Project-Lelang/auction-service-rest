package api

import (
	"context"
	"net/http"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/use_case"
	"auction-service/util"

	"github.com/gin-gonic/gin"
)

type OwnApi struct {
	*api
	productUseCase           use_case.ProductUseCase
	userUseCase              use_case.UserUseCase
	userAddressUseCase       use_case.UserAddressUseCase
	roleRequestUseCase       use_case.RoleRequestUseCase
	withdrawalRequestUseCase use_case.WithdrawalRequestUseCase
	auctionUseCase           use_case.AuctionUseCase
	paymentUseCase           use_case.PaymentUseCase
	bidUseCase               use_case.BidUseCase
	notificationUseCase      use_case.NotificationUseCase
}

// Create godoc
//
//	@Router		/own/products [post]
//	@Summary	Create a new product listing
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnProductCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Create() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder, constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnProductCreateRequest
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
	return a.AuthorizeRoles([]string{constant.RoleBidder, constant.RoleSeller}, func(ctx apiContext) {
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
	return a.AuthorizeRoles([]string{constant.RoleBidder, constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnProductFetchStatusHistoriesRequest
		request.ProductId = ctx.mustGetParamInt64("productId")

		histories := a.productUseCase.OwnFetchStatusHistories(ctx.context(), request)

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
//	@Param		productId	path	int	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Get() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder, constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnProductGetRequest
		request.ProductId = ctx.mustGetParamInt64("productId")

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
//	@Param		productId	path	int									true	"Product ID"
//	@Param		body		body	dto_request.OwnProductUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Update() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder, constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnProductUpdateRequest
		ctx.mustBind(&request)
		request.ProductId = ctx.mustGetParamInt64("productId")

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
//	@Param		productId	path	int	true	"Product ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{product=dto_response.ProductResponse}}
func (a *OwnApi) Request() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder, constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnProductRequestRequest
		request.ProductId = ctx.mustGetParamInt64("productId")

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
	return a.Authorize(func(ctx apiContext) {
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
	return a.Authorize(func(ctx apiContext) {
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

// CreateUserAddress godoc
//
//	@Router		/own/user-addresses [post]
//	@Summary	Create a new user address for the authenticated user
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.UserAddressCreateRequest	true	"Body Request"
//	@Produce	json
//	@Success	201	{object}	dto_response.Response{data=dto_response.DataResponse{user_address=dto_response.UserAddressResponse}}
func (a *OwnApi) CreateUserAddress() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.UserAddressCreateRequest
		ctx.mustBind(&request)

		address := a.userAddressUseCase.OwnCreate(ctx.context(), request)

		ctx.json(http.StatusCreated, dto_response.Response{
			Data: dto_response.DataResponse{
				"user_address": dto_response.NewUserAddressResponse(ctx.context(), address),
			},
		})
	})
}

// FetchUserAddresses godoc
//
//	@Router		/own/user-addresses/filter [post]
//	@Summary	List authenticated user's addresses (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.UserAddressFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.UserAddressResponse}}
func (a *OwnApi) FetchUserAddresses() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.UserAddressFetchRequest
		ctx.mustBind(&request)

		addresses, total := a.userAddressUseCase.OwnFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), addresses, dto_response.NewUserAddressResponse),
				total,
				request.Page,
				request.Limit,
			),
		})
	})
}

// UpdateUserAddress godoc
//
//	@Router		/own/user-addresses/{userAddressId} [put]
//	@Summary	Update authenticated user's address
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		userAddressId	path	int										true	"User Address ID"
//	@Param		body			body	dto_request.UserAddressUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{user_address=dto_response.UserAddressResponse}}
func (a *OwnApi) UpdateUserAddress() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.UserAddressUpdateRequest
		ctx.mustBind(&request)
		request.UserAddressId = ctx.mustGetParamInt64("userAddressId")

		address := a.userAddressUseCase.OwnUpdate(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"user_address": dto_response.NewUserAddressResponse(ctx.context(), address),
			},
		})
	})
}

// DeleteUserAddress godoc
//
//	@Router		/own/user-addresses/{userAddressId} [delete]
//	@Summary	Delete authenticated user's address
//	@tags		Own
//	@Security	BearerAuth
//	@Param		userAddressId	path	int	true	"User Address ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.SuccessResponse}
func (a *OwnApi) DeleteUserAddress() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		request := dto_request.UserAddressDeleteRequest{UserAddressId: ctx.mustGetParamInt64("userAddressId")}

		a.userAddressUseCase.OwnDelete(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.SuccessResponse{Message: "deleted"},
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

// FetchWithdrawalRequests godoc
//
//	@Router		/own/withdrawal-requests/filter [post]
//	@Summary	Get authenticated user's withdrawal request history (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.WithdrawalRequestFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.WithdrawalRequestResponse}}
func (a *OwnApi) FetchWithdrawalRequests() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.WithdrawalRequestFetchRequest
		ctx.mustBind(&request)

		withdrawalRequests, total := a.withdrawalRequestUseCase.OwnFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), withdrawalRequests, func(_ context.Context, wr model.WithdrawalRequest) dto_response.WithdrawalRequestResponse {
					return dto_response.NewWithdrawalRequestResponse(wr)
				}),
				int(total),
				request.Page,
				request.Limit,
			),
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
//	@Param		auctionId	path	int	true	"Auction ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) GetAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionGetRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

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
//	@Param		auctionId	path	int									true	"Auction ID"
//	@Param		body		body	dto_request.OwnAuctionUpdateRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) UpdateAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionUpdateRequest
		ctx.mustBind(&request)
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		auction := a.auctionUseCase.OwnUpdate(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

// FetchBids godoc
//
//	@Router		/own/bids/filter [post]
//	@Summary	Get the authenticated user's own bids (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnBidFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.AuctionBidResponse}}
func (a *OwnApi) FetchBids() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnBidFetchRequest
		ctx.mustBind(&request)

		bids, total := a.bidUseCase.OwnFetch(ctx.context(), request)

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

// GetBid godoc
//
//	@Router		/own/bids/{bidId} [get]
//	@Summary	Get the authenticated user's own bid by ID
//	@tags		Own
//	@Security	BearerAuth
//	@Param		bidId	path	int	true	"Bid ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{bid=dto_response.AuctionBidResponse}}
func (a *OwnApi) GetBid() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnBidGetRequest
		request.BidId = ctx.mustGetParamInt64("bidId")

		bid := a.bidUseCase.OwnGet(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"bid": dto_response.NewAuctionBidResponse(ctx.context(), bid),
			},
		})
	})
}

// FetchPayments godoc
//
//	@Router		/own/payments/filter [post]
//	@Summary	Get the authenticated user's own payments (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnPaymentFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.PaymentResponse}}
func (a *OwnApi) FetchPayments() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnPaymentFetchRequest
		ctx.mustBind(&request)

		payments, total := a.paymentUseCase.OwnFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), payments, dto_response.NewPaymentResponse),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// GetPayment godoc
//
//	@Router		/own/payments/{paymentId} [get]
//	@Summary	Get the authenticated user's own payment by ID
//	@tags		Own
//	@Security	BearerAuth
//	@Param		paymentId	path	int	true	"Payment ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{payment=dto_response.PaymentResponse}}
func (a *OwnApi) GetPayment() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleBidder}, func(ctx apiContext) {
		var request dto_request.OwnPaymentGetRequest
		request.PaymentId = ctx.mustGetParamInt64("paymentId")

		payment := a.paymentUseCase.OwnGet(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"payment": dto_response.NewPaymentResponse(ctx.context(), payment),
			},
		})
	})
}

// RelistAuction godoc
//
//	@Router		/own/auctions/{auctionId}/relist [patch]
//	@Summary	Cancel the auction and relist the product after winner did not pay
//	@tags		Own
//	@Security	BearerAuth
//	@Produce	json
//	@Param		auctionId	path		int	true	"Auction ID"
//	@Success	200			{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) RelistAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionRelistRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		auction := a.auctionUseCase.OwnRelist(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

// SecondChanceAuction godoc
//
//	@Router		/own/auctions/{auctionId}/second-chance [patch]
//	@Summary	Offer the auction to the next-highest bidder after winner did not pay
//	@tags		Own
//	@Security	BearerAuth
//	@Produce	json
//	@Param		auctionId	path		int	true	"Auction ID"
//	@Success	200			{object}	dto_response.Response{data=dto_response.DataResponse{auction=dto_response.AuctionResponse}}
func (a *OwnApi) SecondChanceAuction() gin.HandlerFunc {
	return a.AuthorizeRoles([]string{constant.RoleSeller}, func(ctx apiContext) {
		var request dto_request.OwnAuctionSecondChanceRequest
		request.AuctionId = ctx.mustGetParamInt64("auctionId")

		auction := a.auctionUseCase.OwnSecondChance(ctx.context(), request)
		// Create the initial payment for the new winner (same as post-close flow).
		if err := a.paymentUseCase.CreateInitialPaymentForWinner(ctx.context(), auction.Id); err != nil {
			panic(err)
		}

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"auction": dto_response.NewAuctionResponse(ctx.context(), auction),
			},
		})
	})
}

// FetchNotifications godoc
//
//	@Router		/own/notifications/filter [post]
//	@Summary	Get authenticated user's notifications (paginated)
//	@tags		Own
//	@Security	BearerAuth
//	@Accept		json
//	@Param		body	body	dto_request.OwnNotificationFetchRequest	true	"Body Request"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.PaginationResponse{nodes=[]dto_response.NotificationResponse}}
func (a *OwnApi) FetchNotifications() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		var request dto_request.OwnNotificationFetchRequest
		ctx.mustBind(&request)

		notifications, total := a.notificationUseCase.OwnFetch(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.NewPaginationResponse(
				util.ConvertArray(ctx.context(), notifications, func(_ context.Context, n model.Notification) dto_response.NotificationResponse {
					return dto_response.NewNotificationResponse(n)
				}),
				int(total),
				request.Page,
				request.Limit,
			),
		})
	})
}

// GetNotification godoc
//
//	@Router		/own/notifications/{notificationId} [get]
//	@Summary	Get authenticated user's notification by ID
//	@tags		Own
//	@Security	BearerAuth
//	@Param		notificationId	path	int	true	"Notification ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{notification=dto_response.NotificationResponse}}
func (a *OwnApi) GetNotification() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		request := dto_request.OwnNotificationGetRequest{
			NotificationId: ctx.mustGetParamInt64("notificationId"),
		}

		notification := a.notificationUseCase.OwnGet(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"notification": dto_response.NewNotificationResponse(notification),
			},
		})
	})
}

// MarkNotificationRead godoc
//
//	@Router		/own/notifications/{notificationId}/read [patch]
//	@Summary	Mark authenticated user's notification as read
//	@tags		Own
//	@Security	BearerAuth
//	@Param		notificationId	path	int	true	"Notification ID"
//	@Produce	json
//	@Success	200	{object}	dto_response.Response{data=dto_response.DataResponse{notification=dto_response.NotificationResponse}}
func (a *OwnApi) MarkNotificationRead() gin.HandlerFunc {
	return a.Authorize(func(ctx apiContext) {
		request := dto_request.OwnNotificationMarkReadRequest{
			NotificationId: ctx.mustGetParamInt64("notificationId"),
		}

		notification := a.notificationUseCase.OwnMarkRead(ctx.context(), request)

		ctx.json(http.StatusOK, dto_response.Response{
			Data: dto_response.DataResponse{
				"notification": dto_response.NewNotificationResponse(notification),
			},
		})
	})
}

func RegisterOwnApi(router gin.IRouter, baseApi *api, useCaseManager use_case.UseCaseManager) {
	api := &OwnApi{
		api:                      baseApi,
		productUseCase:           useCaseManager.ProductUseCase(),
		userUseCase:              useCaseManager.UserUseCase(),
		userAddressUseCase:       useCaseManager.UserAddressUseCase(),
		roleRequestUseCase:       useCaseManager.RoleRequestUseCase(),
		withdrawalRequestUseCase: useCaseManager.WithdrawalRequestUseCase(),
		auctionUseCase:           useCaseManager.AuctionUseCase(),
		paymentUseCase:           useCaseManager.PaymentUseCase(),
		bidUseCase:               useCaseManager.BidUseCase(),
		notificationUseCase:      useCaseManager.NotificationUseCase(),
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
	routerGroup.POST("/withdrawal-requests/filter", api.FetchWithdrawalRequests())

	routerAuctionGroup := routerGroup.Group("/auctions")
	routerAuctionGroup.POST("/filter", api.FetchAuctions())
	routerAuctionGroup.GET("/:auctionId", api.GetAuction())
	routerAuctionGroup.POST("", api.CreateAuction())
	routerAuctionGroup.PUT("/:auctionId", api.UpdateAuction())
	routerAuctionGroup.PATCH("/:auctionId/relist", api.RelistAuction())
	routerAuctionGroup.PATCH("/:auctionId/second-chance", api.SecondChanceAuction())

	routerBidGroup := routerGroup.Group("/bids")
	routerBidGroup.POST("/filter", api.FetchBids())
	routerBidGroup.GET("/:bidId", api.GetBid())

	routerPaymentGroup := routerGroup.Group("/payments")
	routerPaymentGroup.POST("/filter", api.FetchPayments())
	routerPaymentGroup.GET("/:paymentId", api.GetPayment())

	routerNotificationGroup := routerGroup.Group("/notifications")
	routerNotificationGroup.POST("/filter", api.FetchNotifications())
	routerNotificationGroup.GET("/:notificationId", api.GetNotification())
	routerNotificationGroup.PATCH("/:notificationId/read", api.MarkNotificationRead())

	// own user-addresses
	routerUserAddressGroup := routerGroup.Group("/user-addresses")
	routerUserAddressGroup.POST("", api.CreateUserAddress())
	routerUserAddressGroup.POST("/filter", api.FetchUserAddresses())
	routerUserAddressGroup.PUT("/:userAddressId", api.UpdateUserAddress())
	routerUserAddressGroup.DELETE("/:userAddressId", api.DeleteUserAddress())
}
