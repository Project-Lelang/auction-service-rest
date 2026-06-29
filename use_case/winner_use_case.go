package use_case

import (
	"context"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

func mustLoadAuctionWinnerData(_ context.Context, repositoryManager repository.RepositoryManager, auctionWinners []*model.AuctionWinner) {
	bidLoader := loader.NewBidLoader(
		repositoryManager.AuctionBidRepository(),
	)

	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, auctionWinner := range auctionWinners {
			group.Go(bidLoader.WinnerFn(auctionWinner))
		}
	}))
}

// WinnerUseCase manages auction winner operations.
type WinnerUseCase interface {
	FetchByAuction(ctx context.Context, request dto_request.AuctionWinnerFetchRequest) ([]model.AuctionWinner, int64)
	GetByAuction(ctx context.Context, request dto_request.AuctionWinnerGetRequest) model.AuctionWinner
}

type winnerUseCase struct {
	repositoryManager repository.RepositoryManager
}

func NewWinnerUseCase(repositoryManager repository.RepositoryManager) WinnerUseCase {
	return &winnerUseCase{repositoryManager: repositoryManager}
}

func (u *winnerUseCase) FetchByAuction(ctx context.Context, request dto_request.AuctionWinnerFetchRequest) ([]model.AuctionWinner, int64) {
	userClaims := model.MustGetUserCtx(ctx)

	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)

	// Only the auction owner (seller) or a superadmin may list winners
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId && !userClaims.HasRole(constant.RoleSuperAdmin) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	option := model.AuctionWinnerQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(
			request.Page,
			request.Limit,
			model.Sorts(request.Sorts),
		),
		AuctionId: util.Pointer(request.AuctionId),
	}

	total, err := u.repositoryManager.AuctionWinnerRepository().Count(ctx, option)
	panicIfErr(err)

	winners, err := u.repositoryManager.AuctionWinnerRepository().Fetch(ctx, option)
	panicIfErr(err)

	winnerPointers := make([]*model.AuctionWinner, len(winners))
	for i := range winners {
		winnerPointers[i] = &winners[i]
	}
	mustLoadAuctionWinnerData(ctx, u.repositoryManager, winnerPointers)

	return winners, total
}

func (u *winnerUseCase) GetByAuction(ctx context.Context, request dto_request.AuctionWinnerGetRequest) model.AuctionWinner {
	userClaims := model.MustGetUserCtx(ctx)

	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)
	winner := mustGetAuctionWinner(ctx, u.repositoryManager, request.WinnerId)

	if winner.AuctionId != auction.Id {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageWinnerNotFound))
	}

	// The winner's bid holds the user_id of the actual buyer
	if winner.AuctionBidId == nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidNotFound))
	}
	bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, *winner.AuctionBidId)
	panicIfRepositoryError(err, constant.LanguageBidNotFound)

	// Only the seller, the buyer themselves, or superadmin may view a winner
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId &&
		bid.UserId != userClaims.UserId &&
		!userClaims.HasRole(constant.RoleSuperAdmin) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	winner.Auction = &auction
	winner.AuctionBid = bid
	return winner
}
