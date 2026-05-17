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

// BidUseCase covers bid read operations for the own (bidder) scope.
type BidUseCase interface {
	OwnFetch(ctx context.Context, request dto_request.OwnBidFetchRequest) ([]model.AuctionBid, int64)
	OwnGet(ctx context.Context, request dto_request.OwnBidGetRequest) model.AuctionBid
}

type bidUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewBidUseCase(repositoryManager repository.RepositoryManager) BidUseCase {
	return &bidUseCase{repositoryManager: repositoryManager}
}

func (u *bidUseCase) OwnFetch(ctx context.Context, request dto_request.OwnBidFetchRequest) ([]model.AuctionBid, int64) {
	userClaims := model.MustGetUserCtx(ctx)

	option := model.AuctionBidQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(
			request.Page,
			request.Limit,
			model.Sorts(request.Sorts),
		),
		UserId:    util.Pointer(userClaims.UserId),
		AuctionId: request.AuctionId,
	}

	total, err := u.repositoryManager.AuctionBidRepository().Count(ctx, option)
	panicIfErr(err)

	bids, err := u.repositoryManager.AuctionBidRepository().Fetch(ctx, option)
	panicIfErr(err)

	return bids, total
}

func (u *bidUseCase) OwnGet(ctx context.Context, request dto_request.OwnBidGetRequest) model.AuctionBid {
	userClaims := model.MustGetUserCtx(ctx)

	bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, request.BidId)
	panicIfRepositoryError(err, constant.LanguageBidNotFound)

	if bid.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	auction := mustGetAuction(ctx, u.repositoryManager, bid.AuctionId)
	bid.Auction = &auction

	return *bid
}
