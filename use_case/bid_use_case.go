package use_case

import (
	"context"
	"fmt"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/internal/notification"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

// BidUseCase covers bid operations for the own (bidder) scope.
type BidUseCase interface {
	OwnFetch(ctx context.Context, request dto_request.OwnBidFetchRequest) ([]model.AuctionBid, int64)
	OwnGet(ctx context.Context, request dto_request.OwnBidGetRequest) model.AuctionBid
	PlaceBid(ctx context.Context, request dto_request.AuctionBidCreateRequest) model.AuctionBid
	PlaceBidNoLocking(ctx context.Context, request dto_request.AuctionBidCreateRequest) model.AuctionBid
}

type bidUseCase struct {
	repositoryManager repository.RepositoryManager
	notificationQueue NotificationPublisher
}

func NewBidUseCase(repositoryManager repository.RepositoryManager, notificationQueue NotificationPublisher) BidUseCase {
	return &bidUseCase{repositoryManager: repositoryManager, notificationQueue: notificationQueue}
}

func ensureAuctionOpenForBid(auction *model.Auction) {
	if auction.Status != constant.AuctionStatusOnGoing ||
		auction.EndTime.IsLessThanOrEqual(util.CurrentDateTime()) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidAuctionNotOnGoing))
	}
}

// mustLoadOwnBidData loads winner and payment for own bids
func (u *bidUseCase) mustLoadOwnBidData(ctx context.Context, bids []*model.AuctionBid) {
	// First, get all auction IDs from bids
	auctionIds := make([]int64, 0, len(bids))
	bidByAuctionId := make(map[int64]*model.AuctionBid)
	for _, bid := range bids {
		auctionIds = append(auctionIds, bid.AuctionId)
		bidByAuctionId[bid.AuctionId] = bid
	}

	if len(auctionIds) == 0 {
		return
	}

	// Fetch winners by auction IDs
	winners, err := u.repositoryManager.AuctionWinnerRepository().FetchByAuctionIds(ctx, auctionIds)
	panicIfErr(err)

	// Map winners to bids and collect payment auction IDs
	winningBidIds := make(map[int64]bool)
	auctionIdsForPayment := make([]int64, 0)
	for _, winner := range winners {
		if bid, ok := bidByAuctionId[winner.AuctionId]; ok {
			if winner.AuctionBidId != nil && bid.Id == *winner.AuctionBidId {
				bid.Winner = &winner
				winningBidIds[bid.Id] = true
				auctionIdsForPayment = append(auctionIdsForPayment, winner.AuctionId)
			}
		}
	}

	// Fetch payments for winning bids
	if len(auctionIdsForPayment) > 0 {
		payments, err := u.repositoryManager.PaymentRepository().FetchByAuctionIds(ctx, auctionIdsForPayment)
		panicIfErr(err)

		paymentByAuctionId := make(map[int64]*model.Payment)
		for i := range payments {
			paymentByAuctionId[payments[i].AuctionId] = &payments[i]
		}

		// Attach payments to winning bids
		for _, bid := range bids {
			if winningBidIds[bid.Id] {
				if payment, ok := paymentByAuctionId[bid.AuctionId]; ok {
					bid.Payment = payment
				}
			}
		}
	}
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

	// Convert to pointer slice for loader
	bidPointers := make([]*model.AuctionBid, len(bids))
	for i := range bids {
		bidPointers[i] = &bids[i]
	}
	u.mustLoadOwnBidData(ctx, bidPointers)

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

	u.mustLoadOwnBidData(ctx, []*model.AuctionBid{bid})

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
	ensureAuctionOpenForBid(&auction)

	// Ensure bidder is not the product owner
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId == userClaims.UserId {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidCannotBidOwnAuction))
	}

	var createdBid model.AuctionBid
	var outbidUserId int64

	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		// Lock the auction row to serialise concurrent bids
		auction, err := u.repositoryManager.AuctionRepository().GetById(ctx, request.AuctionId)
		if err != nil {
			return err
		}
		ensureAuctionOpenForBid(auction)

		// Lock the current active winner row (gap lock if none exists)
		currentWinner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, request.AuctionId)
		if err != nil {
			return err
		}
		mustLoadAuctionWinnerData(ctx, u.repositoryManager, []*model.AuctionWinner{currentWinner})

		minAmount := auction.StartingPrice
		if currentWinner.AuctionBid != nil {
			minAmount = currentWinner.AuctionBid.Amount
			if currentWinner.AuctionBid.UserId != userClaims.UserId {
				outbidUserId = currentWinner.AuctionBid.UserId
			}
		}
		if request.Amount <= minAmount {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidAmountTooLow))
		}

		// Insert the new bid
		bid := model.AuctionBid{
			UserId:    userClaims.UserId,
			AuctionId: request.AuctionId,
			Amount:    request.Amount,
		}
		if err := u.repositoryManager.AuctionBidRepository().Insert(ctx, &bid); err != nil {
			return err
		}
		createdBid = bid

		// Update the existing winner to point to the new highest bid
		_, err = u.repositoryManager.AuctionWinnerRepository().UpdateBidId(ctx, currentWinner.Id, bid.Id)
		return err
	})
	panicIfErr(err)

	createdBid.Auction = &auction
	if outbidUserId != 0 {
		publishAuctionNotification(ctx, u.notificationQueue, notification.Payload{
			UserId:    outbidUserId,
			Role:      notification.RoleBuyer,
			EventType: notification.EventOutbid,
			AuctionId: auction.Id,
			Title:     "Outbid!",
			Body:      "Your bid has been outbid.",
			DataPayload: map[string]string{
				"auction_url":         auctionURL(auction.Id),
				"current_highest_bid": fmt.Sprintf("%.0f", createdBid.Amount),
			},
		})
	}
	return createdBid
}

func (u *bidUseCase) PlaceBidNoLocking(ctx context.Context, request dto_request.AuctionBidCreateRequest) model.AuctionBid {
	userClaims := model.MustGetUserCtx(ctx)

	if !userClaims.HasRole(constant.RoleBidder) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)
	ensureAuctionOpenForBid(&auction)

	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId == userClaims.UserId {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidCannotBidOwnAuction))
	}

	var createdBid model.AuctionBid
	var outbidUserId int64

	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		auction, err := u.repositoryManager.AuctionRepository().GetById(ctx, request.AuctionId)
		if err != nil {
			return err
		}
		ensureAuctionOpenForBid(auction)

		currentWinner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdNoLocking(ctx, request.AuctionId)
		if err != nil {
			return err
		}
		mustLoadAuctionWinnerData(ctx, u.repositoryManager, []*model.AuctionWinner{currentWinner})

		minAmount := auction.StartingPrice
		if currentWinner.AuctionBid != nil {
			minAmount = currentWinner.AuctionBid.Amount
			if currentWinner.AuctionBid.UserId != userClaims.UserId {
				outbidUserId = currentWinner.AuctionBid.UserId
			}
		}
		if request.Amount <= minAmount {
			panic(dto_response.NewBadRequestErrorResponse(constant.LanguageBidAmountTooLow))
		}

		bid := model.AuctionBid{
			UserId:    userClaims.UserId,
			AuctionId: request.AuctionId,
			Amount:    request.Amount,
		}
		if err := u.repositoryManager.AuctionBidRepository().Insert(ctx, &bid); err != nil {
			return err
		}
		createdBid = bid

		_, err = u.repositoryManager.AuctionWinnerRepository().UpdateBidId(ctx, currentWinner.Id, bid.Id)
		return err
	})
	panicIfErr(err)

	createdBid.Auction = &auction
	if outbidUserId != 0 {
		publishAuctionNotification(ctx, u.notificationQueue, notification.Payload{
			UserId:    outbidUserId,
			Role:      notification.RoleBuyer,
			EventType: notification.EventOutbid,
			AuctionId: auction.Id,
			Title:     "Outbid!",
			Body:      "Your bid has been outbid.",
			DataPayload: map[string]string{
				"auction_url":         auctionURL(auction.Id),
				"current_highest_bid": fmt.Sprintf("%.0f", createdBid.Amount),
			},
		})
	}
	return createdBid
}
