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

// BidUseCase covers bid operations for the own (bidder) scope.
type BidUseCase interface {
	OwnFetch(ctx context.Context, request dto_request.OwnBidFetchRequest) ([]model.AuctionBid, int64)
	OwnGet(ctx context.Context, request dto_request.OwnBidGetRequest) model.AuctionBid
	PlaceBid(ctx context.Context, request dto_request.AuctionBidCreateRequest) model.AuctionBid
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

// PlaceBid places a new bid on an ON_GOING auction.
// Pessimistic concurrency control is achieved by locking the auction row and
// the current active winner row inside a transaction before inserting/updating.
func (u *bidUseCase) PlaceBid(ctx context.Context, request dto_request.AuctionBidCreateRequest) model.AuctionBid {
	userClaims := model.MustGetUserCtx(ctx)

	if !userClaims.HasRole(constant.RoleBidder) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// Pre-check auction status (no lock yet – just an early guard)
	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)
	if auction.Status != constant.AuctionStatusOnGoing {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidAuctionNotOnGoing))
	}

	// Ensure bidder is not the product owner
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId == userClaims.UserId {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidCannotBidOwnAuction))
	}

	var createdBid model.AuctionBid

	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		// Lock the auction row to serialise concurrent bids
		lockedAuction, err := u.repositoryManager.AuctionRepository().GetByIdForUpdate(ctx, request.AuctionId)
		if err != nil {
			return err
		}
		if lockedAuction.Status != constant.AuctionStatusOnGoing {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidAuctionNotOnGoing))
		}

		// Determine the minimum valid bid amount
		highestBid, err := u.repositoryManager.AuctionBidRepository().GetHighestByAuctionId(ctx, request.AuctionId)
		if err != nil && err != constant.ErrNoData {
			return err
		}

		minAmount := lockedAuction.StartingPrice
		if highestBid != nil {
			minAmount = highestBid.Amount
		}
		if request.Amount <= minAmount {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidAmountTooLow))
		}

		// Insert the new bid
		bid := model.AuctionBid{
			Id:        util.NewUuid(),
			UserId:    userClaims.UserId,
			AuctionId: request.AuctionId,
			Amount:    request.Amount,
		}
		if err := u.repositoryManager.AuctionBidRepository().Insert(ctx, &bid); err != nil {
			return err
		}
		createdBid = bid

		// Lock the current active winner row (gap lock if none exists)
		currentWinner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, request.AuctionId)
		if err == constant.ErrNoData {
			// No winner yet – create the first one
			winner := model.AuctionWinner{
				Id:           util.NewUuid(),
				AuctionId:    request.AuctionId,
				AuctionBidId: bid.Id,
				Status:       constant.AuctionWinnerStatusOnGoing,
			}
			return u.repositoryManager.AuctionWinnerRepository().Insert(ctx, &winner)
		}
		if err != nil {
			return err
		}

		// Update the existing winner to point to the new highest bid
		_, err = u.repositoryManager.AuctionWinnerRepository().UpdateBidId(ctx, currentWinner.Id, bid.Id)
		return err
	})
	panicIfErr(err)

	createdBid.Auction = &auction
	return createdBid
}
