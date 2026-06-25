package use_case

import (
	"context"
	"log"
	"strconv"
	"time"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	internalFilesystem "auction-service/internal/filesystem"
	"auction-service/internal/notification"
	"auction-service/loader"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"

	"golang.org/x/sync/errgroup"
)

// TaskQueue abstracts delayed auction task scheduling.
// Implemented by infrastructure.TaskQueueClient (duck typing — no import needed).
type TaskQueue interface {
	EnqueueAuctionStart(auctionId int64, processAt time.Time) error
	EnqueueAuctionClose(auctionId int64, processAt time.Time) error
	EnqueuePaymentExpiry(paymentId int64, processAt time.Time) error
	EnqueueShipmentAddressDue(shipmentId int64, processAt time.Time) error
	EnqueueShipmentShipDue(shipmentId int64, processAt time.Time) error
	EnqueueShipmentTrackCheck(shipmentId int64, processAt time.Time) error
	EnqueueShipmentReceiveDue(shipmentId int64, processAt time.Time) error
	ReplaceAuctionStart(auctionId int64, processAt time.Time) error
}

// AuctionUseCase covers all auction operations for the own (seller) scope
// and public viewing endpoints.
type AuctionUseCase interface {
	// public
	Fetch(ctx context.Context, request dto_request.AuctionFetchRequest) ([]model.Auction, int64)
	Get(ctx context.Context, request dto_request.AuctionGetRequest) model.Auction
	GetAdminDashboardReport(ctx context.Context, request dto_request.AdminDashboardReportRequest) []dto_response.DashboardDailyReport

	// own (seller)
	OwnFetch(ctx context.Context, request dto_request.OwnAuctionFetchRequest) ([]model.Auction, int64)
	OwnGet(ctx context.Context, request dto_request.OwnAuctionGetRequest) model.Auction
	OwnCreate(ctx context.Context, request dto_request.OwnAuctionCreateRequest) model.Auction
	OwnUpdate(ctx context.Context, request dto_request.OwnAuctionUpdateRequest) model.Auction

	// own seller-decision after winner failed to pay
	OwnRelist(ctx context.Context, request dto_request.OwnAuctionRelistRequest) model.Auction
	OwnSecondChance(ctx context.Context, request dto_request.OwnAuctionSecondChanceRequest) model.Auction

	// task handlers — called by the asynq worker goroutine
	HandleStartAuction(ctx context.Context, auctionId int64) error
	HandleCloseAuction(ctx context.Context, auctionId int64) error

	// startup recovery: re-enqueue tasks for SCHEDULED / ON_GOING auctions
	// Returns IDs of auctions that were directly closed and need payment init.
	EnqueueScheduledTasks(ctx context.Context) []int64
}

type auctionUseCase struct {
	repositoryManager repository.RepositoryManager
	filesystemManager internalFilesystem.FilesystemManager
	taskQueue         TaskQueue
	notificationQueue NotificationPublisher
}

func NewAuctionUseCase(repositoryManager repository.RepositoryManager, filesystemManager internalFilesystem.FilesystemManager, taskQueue TaskQueue, notificationQueue NotificationPublisher) AuctionUseCase {
	return &auctionUseCase{
		repositoryManager: repositoryManager,
		filesystemManager: filesystemManager,
		taskQueue:         taskQueue,
		notificationQueue: notificationQueue,
	}
}

func (u *auctionUseCase) populateAuctionProductImageLinks(auctions ...*model.Auction) {
	const presignedExpiry = 24 * time.Hour
	mainFs := u.filesystemManager.Main()

	for _, auction := range auctions {
		if auction.Product == nil {
			continue
		}
		product := auction.Product
		if product.CoverImagePath != nil && *product.CoverImagePath != "" {
			link := mainFs.PresignedUrl(util.GetFilenameFromPath(*product.CoverImagePath), *product.CoverImagePath, presignedExpiry)
			product.CoverImageLink = &link
		}

		imagePaths := model.ParseImagePaths(product.ImagePaths)
		product.ImageLinks = make([]string, 0, len(imagePaths))
		for _, p := range imagePaths {
			product.ImageLinks = append(product.ImageLinks, mainFs.PresignedUrl(util.GetFilenameFromPath(p), p, presignedExpiry))
		}
	}
}

// mustLoadAuctionData loads product and winner for public auction endpoints (no payment)
func (u *auctionUseCase) mustLoadAuctionData(_ context.Context, auctions []*model.Auction) {
	productLoader := loader.NewProductLoader(u.repositoryManager.ProductRepository())
	winnerLoader := loader.NewWinnerLoader(
		u.repositoryManager.AuctionWinnerRepository(),
		u.repositoryManager.AuctionBidRepository(),
		u.repositoryManager.UserRepository(),
	)

	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, auction := range auctions {
			group.Go(productLoader.AuctionFn(auction))
			group.Go(winnerLoader.AuctionFn(auction))
		}
	}))
	u.populateAuctionProductImageLinks(auctions...)
}

// mustLoadOwnAuctionData loads product, winner, payment, and bids for own auction endpoints
func (u *auctionUseCase) mustLoadOwnAuctionData(_ context.Context, auctions []*model.Auction) {
	productLoader := loader.NewProductLoader(u.repositoryManager.ProductRepository())
	winnerLoader := loader.NewWinnerLoader(
		u.repositoryManager.AuctionWinnerRepository(),
		u.repositoryManager.AuctionBidRepository(),
		u.repositoryManager.UserRepository(),
	)
	paymentLoader := loader.NewPaymentLoader(u.repositoryManager.PaymentRepository())
	bidsLoader := loader.NewAuctionBidsLoader(
		u.repositoryManager.AuctionBidRepository(),
		u.repositoryManager.UserRepository(),
	)

	panicIfErr(util.Await(func(group *errgroup.Group) {
		for _, auction := range auctions {
			group.Go(productLoader.AuctionFn(auction))
			group.Go(winnerLoader.AuctionFn(auction))
			group.Go(paymentLoader.AuctionFn(auction))
			group.Go(bidsLoader.AuctionFn(auction))
		}
	}))
	u.populateAuctionProductImageLinks(auctions...)
}

func (u *auctionUseCase) mustGetOwnAuction(ctx context.Context, auctionId int64, userId int64) model.Auction {
	auction := mustGetAuction(ctx, u.repositoryManager, auctionId)
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}
	auction.Product = &product
	return auction
}

func (u *auctionUseCase) Fetch(ctx context.Context, request dto_request.AuctionFetchRequest) ([]model.Auction, int64) {
	option := model.AuctionQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(
			request.Page,
			request.Limit,
			model.Sorts(request.Sorts),
		),
		Status: request.Status,
	}

	total, err := u.repositoryManager.AuctionRepository().Count(ctx, option)
	panicIfErr(err)

	auctions, err := u.repositoryManager.AuctionRepository().Fetch(ctx, option)
	panicIfErr(err)

	u.mustLoadAuctionData(ctx, util.SliceValueToSlicePointer(auctions))

	return auctions, total
}

func (u *auctionUseCase) Get(ctx context.Context, request dto_request.AuctionGetRequest) model.Auction {
	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)
	u.mustLoadAuctionData(ctx, []*model.Auction{&auction})
	return auction
}

func (u *auctionUseCase) GetAdminDashboardReport(ctx context.Context, request dto_request.AdminDashboardReportRequest) []dto_response.DashboardDailyReport {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	startDateStr := request.StartDate
	endDateStr := request.EndDate

	// Default ke 30 hari terakhir jika kosong
	if startDateStr == "" {
		startDateStr = now.AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDateStr == "" {
		endDateStr = now.Format("2006-01-02")
	}

	startTime, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
	panicIfErr(err)

	endTime, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
	panicIfErr(err)
	// Pastikan mencakup waktu sampai akhir hari (23:59:59)
	endTime = endTime.Add(24*time.Hour - time.Second)

	// Panggil repository yang baru dibuat
	reports, err := u.repositoryManager.AuctionRepository().GetDailyReport(ctx, startTime, endTime)
	panicIfErr(err)

	return reports
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

	u.mustLoadOwnAuctionData(ctx, util.SliceValueToSlicePointer(auctions))

	return auctions, total
}

func (u *auctionUseCase) OwnGet(ctx context.Context, request dto_request.OwnAuctionGetRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)
	auction := u.mustGetOwnAuction(ctx, request.AuctionId, userClaims.UserId)

	u.mustLoadOwnAuctionData(ctx, []*model.Auction{&auction})

	return auction
}

func (u *auctionUseCase) OwnCreate(ctx context.Context, request dto_request.OwnAuctionCreateRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)
	productId := request.ProductId.Int64()

	// verify user has SELLER role
	if !userClaims.HasRole(constant.RoleSeller) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// validate time range
	if !request.EndTime.IsGreaterThan(request.StartTime) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionInvalidTimeRange))
	}

	// start_time must be at least 1 hour from now
	// if request.StartTime.Time().Before(time.Now().Add(time.Hour)) {
	// 	panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionStartTimeTooSoon))
	// }

	// load product and verify ownership + status
	product := mustGetProduct(ctx, u.repositoryManager, productId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}
	if product.Status != constant.ProductStatusVerified {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionProductNotVerified))
	}

	auction := model.Auction{
		ProductId:     productId,
		StartingPrice: request.StartingPrice,
		StartTime:     request.StartTime,
		EndTime:       request.EndTime,
		Status:        constant.AuctionStatusScheduled,
		Fee:           0,
	}

	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		err := u.repositoryManager.AuctionRepository().Insert(ctx, &auction)
		if err != nil {
			return err
		}

		winner := model.AuctionWinner{
			AuctionId:    auction.Id,
			AuctionBidId: nil,
			Status:       constant.AuctionWinnerStatusOnGoing,
		}
		errWinner := u.repositoryManager.AuctionWinnerRepository().Insert(ctx, &winner)
		if errWinner != nil {
			return errWinner
		}

		// Schedule the start task; EnqueueScheduledTasks on startup recovers if this fails.
		if errRedis := u.taskQueue.EnqueueAuctionStart(auction.Id, auction.StartTime.Time()); errRedis != nil {
			log.Printf("[auction worker] enqueue start for %d failed: %v", auction.Id, errRedis)
			return errRedis
		}

		return nil
	})
	panicIfErr(err)

	auction.Product = &product
	u.populateAuctionProductImageLinks(&auction)
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

	// start_time must be at least 1 hour from now
	if request.StartTime.Time().Before(time.Now().Add(time.Hour)) {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionStartTimeTooSoon))
	}

	updated, err := u.repositoryManager.AuctionRepository().Update(
		ctx,
		auction.Id,
		request.StartingPrice,
		request.StartTime,
		request.EndTime,
		0,
	)
	panicIfErr(err)

	updated.Product = auction.Product
	u.populateAuctionProductImageLinks(updated)

	// Replace the scheduled start task with the new timing.
	if err := u.taskQueue.ReplaceAuctionStart(updated.Id, updated.StartTime.Time()); err != nil {
		log.Printf("[auction worker] replace start task for %d failed: %v", updated.Id, err)
	}

	return *updated
}

// HandleStartAuction is called by the asynq worker when an auction's start_time is reached.
// It transitions the auction SCHEDULED → ON_GOING and the product → ON_BIDS,
// then enqueues the close task.
func (u *auctionUseCase) HandleStartAuction(ctx context.Context, auctionId int64) error {
	auction, err := u.repositoryManager.AuctionRepository().GetById(ctx, auctionId)
	if err != nil {
		return err
	}
	if auction == nil {
		log.Printf("[auction worker] start: auction %d not found, skip", auctionId)
		return nil
	}
	// Idempotency guard: skip if already started.
	if auction.Status != constant.AuctionStatusScheduled {
		log.Printf("[auction worker] start: auction %d already %s, skip", auctionId, auction.Status)
		return nil
	}

	err = u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auctionId, constant.AuctionStatusOnGoing); err != nil {
			return err
		}
		if _, err := u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusOnBids); err != nil {
			return err
		}
		return u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
			ProductId: auction.ProductId,
			Status:    constant.ProductStatusOnBids,
		})
	})
	if err != nil {
		return err
	}

	// Enqueue the close task after the DB transaction commits.
	if err := u.taskQueue.EnqueueAuctionClose(auctionId, auction.EndTime.Time()); err != nil {
		log.Printf("[auction worker] enqueue close for %d failed: %v (will recover on restart)", auctionId, err)
	}
	product, err := u.repositoryManager.ProductRepository().GetById(ctx, auction.ProductId)
	if err == nil && product != nil {
		publishAuctionNotification(ctx, u.notificationQueue, notification.Payload{
			UserId:    product.UserId,
			Role:      notification.RoleSeller,
			EventType: notification.EventAuctionStart,
			AuctionId: auction.Id,
			Title:     "Auction is live",
			Body:      "Your auction is now live.",
			DataPayload: map[string]string{
				"auction_url": auctionURL(auction.Id),
				"product_id":  strconv.FormatInt(auction.ProductId, 10),
			},
		})
	}
	return nil
}

// HandleCloseAuction is called by the asynq worker when an auction's end_time is reached.
func (u *auctionUseCase) HandleCloseAuction(ctx context.Context, auctionId int64) error {
	auction, err := u.repositoryManager.AuctionRepository().GetById(ctx, auctionId)
	if err != nil {
		return err
	}
	if auction == nil {
		log.Printf("[auction worker] close: auction %d not found, skip", auctionId)
		return nil
	}
	// Idempotency guard: skip if not ON_GOING.
	if auction.Status != constant.AuctionStatusOnGoing {
		log.Printf("[auction worker] close: auction %d already %s, skip", auctionId, auction.Status)
		return nil
	}
	return u.closeAuction(ctx, *auction)
}

// EnqueueScheduledTasks is called on server startup to recover in-flight auction state.
// It re-enqueues tasks for SCHEDULED/ON_GOING auctions and directly closes any
// ON_GOING auction whose end_time already passed (to avoid ErrTaskIDConflict with
// stale completed tasks in the Redis archive).
// Returns a slice of auction IDs that were directly closed and need payment init,
// plus any WAITING_FOR_PAYMENT auctions that may be missing their initial payment.
func (u *auctionUseCase) EnqueueScheduledTasks(ctx context.Context) []int64 {
	var needPaymentInit []int64
	now := time.Now()

	scheduledStatus := constant.AuctionStatusScheduled
	scheduled, err := u.repositoryManager.AuctionRepository().Fetch(ctx, model.AuctionQueryOption{Status: &scheduledStatus})
	if err != nil {
		log.Printf("[auction worker] fetch scheduled auctions error: %v", err)
	} else {
		startedDirectly := 0
		for _, a := range scheduled {
			if a.StartTime.Time().Before(now) {
				// Start time already passed — start directly without going through
				// the task queue to avoid ErrTaskIDConflict with stale tasks.
				// After this, the auction becomes ON_GOING and will be picked up
				// by the ON_GOING loop below (handles double-missed start+end case).
				log.Printf("[auction worker] starting overdue auction %d directly", a.Id)
				if err := u.HandleStartAuction(ctx, a.Id); err != nil {
					log.Printf("[auction worker] direct start for %d failed: %v", a.Id, err)
				} else {
					startedDirectly++
				}
			} else {
				if err := u.taskQueue.EnqueueAuctionStart(a.Id, a.StartTime.Time()); err != nil {
					log.Printf("[auction worker] re-enqueue start for %d failed: %v", a.Id, err)
				}
			}
		}
		log.Printf("[auction worker] processed %d scheduled auction(s) on startup (%d started directly)", len(scheduled), startedDirectly)
	}

	onGoingStatus := constant.AuctionStatusOnGoing
	onGoing, err := u.repositoryManager.AuctionRepository().Fetch(ctx, model.AuctionQueryOption{Status: &onGoingStatus})
	if err != nil {
		log.Printf("[auction worker] fetch on-going auctions error: %v", err)
	} else {
		for _, a := range onGoing {
			if a.EndTime.Time().Before(now) {
				// End time already passed — close directly without going through
				// the task queue to avoid ErrTaskIDConflict with stale tasks.
				log.Printf("[auction worker] closing overdue auction %d directly", a.Id)
				if err := u.closeAuction(ctx, a); err != nil {
					log.Printf("[auction worker] direct close for %d failed: %v", a.Id, err)
				} else {
					needPaymentInit = append(needPaymentInit, a.Id)
				}
			} else {
				if err := u.taskQueue.EnqueueAuctionClose(a.Id, a.EndTime.Time()); err != nil {
					log.Printf("[auction worker] re-enqueue close for %d failed: %v", a.Id, err)
				}
			}
		}
		log.Printf("[auction worker] processed %d on-going auction(s) on startup", len(onGoing))
	}

	// Recovery: collect WAITING_FOR_PAYMENT auctions that may have lost their
	// initial payment creation (e.g. server crashed between closeAuction and
	// CreateInitialPaymentForWinner).
	waitingStatus := constant.AuctionStatusWaitingForPayment
	waiting, err := u.repositoryManager.AuctionRepository().Fetch(ctx, model.AuctionQueryOption{Status: &waitingStatus})
	if err != nil {
		log.Printf("[auction worker] fetch waiting-for-payment auctions error: %v", err)
	} else {
		for _, a := range waiting {
			needPaymentInit = append(needPaymentInit, a.Id)
		}
		if len(waiting) > 0 {
			log.Printf("[auction worker] queued payment init recovery for %d waiting-for-payment auction(s)", len(waiting))
		}
	}

	return needPaymentInit
}

// closeAuction contains the shared closing logic used by both the task handler
// and any manual close paths.
func (u *auctionUseCase) closeAuction(ctx context.Context, auction model.Auction) error {
	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		// Lock the active winner row (if any) to decide the closing path and
		// prevent concurrent ticks from processing the same auction twice.
		winner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, auction.Id)
		if err != nil && err != constant.ErrNoData {
			return err
		}

		if winner == nil || winner.AuctionBidId == nil {
			// No bids — cancel auction and revert product to VERIFIED.
			if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusCancelled); err != nil {
				return err
			}
			if winner != nil {
				if _, err := u.repositoryManager.AuctionWinnerRepository().UpdateStatus(ctx, winner.Id, constant.AuctionWinnerStatusCancelled); err != nil {
					return err
				}
			}
			if _, err := u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusVerified); err != nil {
				return err
			}
			msg := "Auction ended with no bids"
			return u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
				ProductId: auction.ProductId,
				Status:    constant.ProductStatusVerified,
				Message:   &msg,
			})
		}

		// Has winner — move auction and product to WAITING_FOR_PAYMENT.
		// The winner record itself was already created (and kept current) by the
		// bid use case, so no new winner insert is needed here.
		if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusWaitingForPayment); err != nil {
			return err
		}
		_, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusWaitingForPayment)
		return err
	})
	if err != nil {
		return err
	}

	product, productErr := u.repositoryManager.ProductRepository().GetById(ctx, auction.ProductId)
	if productErr == nil && product != nil {
		publishAuctionNotification(ctx, u.notificationQueue, notification.Payload{
			UserId:    product.UserId,
			Role:      notification.RoleSeller,
			EventType: notification.EventAuctionEnd,
			AuctionId: auction.Id,
			Title:     "Auction ended",
			Body:      "Your auction has ended.",
			DataPayload: map[string]string{
				"auction_url": auctionURL(auction.Id),
				"product_id":  strconv.FormatInt(auction.ProductId, 10),
			},
		})
	}

	return nil
}

// OwnRelist is called when the seller decides to relist the product after a
// winner did not pay.  The auction is cancelled and the product reverts to VERIFIED.
func (u *auctionUseCase) OwnRelist(ctx context.Context, request dto_request.OwnAuctionRelistRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)
	auction := u.mustGetOwnAuction(ctx, request.AuctionId, userClaims.UserId)

	if auction.Status != constant.AuctionStatusWaitingForSellerDecision {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionNotWaitingForSellerDecision))
	}

	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusCancelled); err != nil {
			return err
		}
		if _, err := u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusVerified); err != nil {
			return err
		}
		msg := "Auction cancelled by seller after winner did not pay"
		return u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
			ProductId: auction.ProductId,
			Status:    constant.ProductStatusVerified,
			Message:   &msg,
		})
	}))

	auction.Status = constant.AuctionStatusCancelled
	u.populateAuctionProductImageLinks(&auction)
	return auction
}

// OwnSecondChance offers the auction to the next-highest bidder after the
// original winner did not pay.  Returns an error if no next bidder exists.
func (u *auctionUseCase) OwnSecondChance(ctx context.Context, request dto_request.OwnAuctionSecondChanceRequest) model.Auction {
	userClaims := model.MustGetUserCtx(ctx)
	auction := u.mustGetOwnAuction(ctx, request.AuctionId, userClaims.UserId)

	if auction.Status != constant.AuctionStatusWaitingForSellerDecision {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionNotWaitingForSellerDecision))
	}

	// Collect user IDs of all previously cancelled winners so we can exclude
	// ALL of their bids from the next-highest-bid query (a user may have placed
	// multiple bids; excluding only the winning bid_id would still let their
	// lower bids surface as the "next highest").
	cancelledStatus := constant.AuctionWinnerStatusCancelled
	cancelledWinners, err := u.repositoryManager.AuctionWinnerRepository().Fetch(ctx, model.AuctionWinnerQueryOption{
		AuctionId: &request.AuctionId,
		Status:    &cancelledStatus,
	})
	panicIfErr(err)

	seenUsers := make(map[int64]struct{}, len(cancelledWinners))
	for _, w := range cancelledWinners {
		if w.AuctionBidId == nil {
			continue
		}
		bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, *w.AuctionBidId)
		panicIfErr(err)
		seenUsers[bid.UserId] = struct{}{}
	}
	excludeUserIds := make([]int64, 0, len(seenUsers))
	for uid := range seenUsers {
		excludeUserIds = append(excludeUserIds, uid)
	}

	nextBid, err := u.repositoryManager.AuctionBidRepository().GetNextHighestByAuctionIdExcludingUsers(ctx, auction.Id, excludeUserIds)
	panicIfErr(err, constant.ErrNoData)
	if err == constant.ErrNoData || nextBid == nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionNoNextBidder))
	}

	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		newWinner := model.AuctionWinner{
			AuctionId:    auction.Id,
			AuctionBidId: &nextBid.Id,
			Status:       constant.AuctionWinnerStatusOnGoing,
		}
		if err := u.repositoryManager.AuctionWinnerRepository().Insert(ctx, &newWinner); err != nil {
			return err
		}
		if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusWaitingForPayment); err != nil {
			return err
		}
		_, err := u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusWaitingForPayment)
		return err
	}))

	auction.Status = constant.AuctionStatusWaitingForPayment
	u.populateAuctionProductImageLinks(&auction)
	return auction
}
