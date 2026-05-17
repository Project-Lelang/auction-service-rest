package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

// AuctionUseCase covers all auction operations for the own (seller) scope.
type AuctionUseCase interface {
	OwnFetch(ctx context.Context, request dto_request.OwnAuctionFetchRequest) ([]model.Auction, int64)
	OwnGet(ctx context.Context, request dto_request.OwnAuctionGetRequest) model.Auction
	OwnCreate(ctx context.Context, request dto_request.OwnAuctionCreateRequest) model.Auction
	OwnUpdate(ctx context.Context, request dto_request.OwnAuctionUpdateRequest) model.Auction
}

type auctionUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewAuctionUseCase(repositoryManager repository.RepositoryManager) AuctionUseCase {
	return &auctionUseCase{repositoryManager: repositoryManager}
}

func (u *auctionUseCase) mustGetOwnAuction(ctx context.Context, auctionId string, userId string) model.Auction {
	auction := mustGetAuction(ctx, u.repositoryManager, auctionId)
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}
	auction.Product = &product
	return auction
}

func (u *auctionUseCase) OwnFetch(ctx context.Context, request dto_request.OwnAuctionFetchRequest) ([]model.Auction, int64) {
	userClaims := model.MustGetUserCtx(ctx)

	option := model.AuctionQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(
			request.Page,
			request.Limit,
			model.Sorts(request.Sorts),
		),
		UserId: util.Pointer(userClaims.UserId),
		Status: request.Status,
	}

	total, err := u.repositoryManager.AuctionRepository().Count(ctx, option)
	panicIfErr(err)

	auctions, err := u.repositoryManager.AuctionRepository().Fetch(ctx, option)
	panicIfErr(err)

	return auctions, total
}

func (u *auctionUseCase) OwnGet(ctx context.Context, request dto_request.OwnAuctionGetRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)
	return u.mustGetOwnAuction(ctx, request.AuctionId, userClaims.UserId)
}

func (u *auctionUseCase) OwnCreate(ctx context.Context, request dto_request.OwnAuctionCreateRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)

	// verify user has SELLER role
	if !userClaims.HasRole(constant.RoleSeller) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// validate time range
	if !request.EndTime.IsGreaterThan(request.StartTime) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionInvalidTimeRange))
	}

	// load product and verify ownership + status
	product := mustGetProduct(ctx, u.repositoryManager, request.ProductId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}
	if product.Status != constant.ProductStatusVerified {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionProductNotVerified))
	}

	auction := model.Auction{
		Id:            util.NewUuid(),
		ProductId:     request.ProductId,
		StartingPrice: request.StartingPrice,
		StartTime:     request.StartTime,
		EndTime:       request.EndTime,
		Status:        constant.AuctionStatusScheduled,
		Fee:           constant.AuctionFee,
	}
	panicIfErr(u.repositoryManager.AuctionRepository().Insert(ctx, &auction))

	// mark product as ON_BIDS
	_, err := u.repositoryManager.ProductRepository().UpdateStatus(ctx, product.Id, constant.ProductStatusOnBids)
	panicIfErr(err)

	auction.Product = &product
	return auction
}

func (u *auctionUseCase) OwnUpdate(ctx context.Context, request dto_request.OwnAuctionUpdateRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)

	auction := u.mustGetOwnAuction(ctx, request.AuctionId, userClaims.UserId)

	// only SCHEDULED auctions may be updated
	if auction.Status != constant.AuctionStatusScheduled {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionNotScheduled))
	}

	// validate time range
	if !request.EndTime.IsGreaterThan(request.StartTime) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionInvalidTimeRange))
	}

	updated, err := u.repositoryManager.AuctionRepository().Update(
		ctx,
		auction.Id,
		request.StartingPrice,
		request.StartTime,
		request.EndTime,
		constant.AuctionFee,
	)
	panicIfErr(err)

	updated.Product = auction.Product
	return *updated
}
