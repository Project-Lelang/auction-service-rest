package use_case

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

// PaymentUseCase manages payment creation and notification handling.
type PaymentUseCase interface {
	GetByAuction(ctx context.Context, request dto_request.AuctionPaymentGetRequest) model.Payment
	OwnFetch(ctx context.Context, request dto_request.OwnPaymentFetchRequest) ([]model.Payment, int64)
	OwnGet(ctx context.Context, request dto_request.OwnPaymentGetRequest) model.Payment
	HandleMidtransNotification(ctx context.Context, notification infrastructure.MidtransNotification)
	// HandlePaymentExpiry is the safety-net task handler called by the asynq worker
	// when a payment's expired_at window elapses without a Midtrans webhook.
	HandlePaymentExpiry(ctx context.Context, paymentId int64) error
	// CreateInitialPaymentForWinner creates the first payment after an auction closes.
	// Called by the asynq close-auction handler after HandleCloseAuction succeeds.
	// Idempotent: no-op if the winner already has a WAITING_FOR_PAYMENT record.
	CreateInitialPaymentForWinner(ctx context.Context, auctionId int64) error
	// RecoverExpiredPayments is called on startup to process any payments that are
	// still WAITING_FOR_PAYMENT but whose expired_at has already passed.
	RecoverExpiredPayments(ctx context.Context) error
}

type paymentUseCase struct {
	repositoryManager repository.RepositoryManager
	midtransClient    infrastructure.MidtransClient
	biteshipClient    infrastructure.BiteshipClient
	taskQueue         TaskQueue
}

func NewPaymentUseCase(repositoryManager repository.RepositoryManager, midtransClient infrastructure.MidtransClient, taskQueue TaskQueue, biteshipClient infrastructure.BiteshipClient, _ NotificationPublisher) PaymentUseCase {
	return &paymentUseCase{
		repositoryManager: repositoryManager,
		midtransClient:    midtransClient,
		biteshipClient:    biteshipClient,
		taskQueue:         taskQueue,
	}
}

func (u *paymentUseCase) GetByAuction(ctx context.Context, request dto_request.AuctionPaymentGetRequest) model.Payment {
	userClaims := model.MustGetUserCtx(ctx)

	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)
	payment := mustGetPayment(ctx, u.repositoryManager, request.PaymentId)

	if payment.AuctionId != auction.Id {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguagePaymentNotFound))
	}

	// Only the seller, the payer, or superadmin may view a payment
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId &&
		payment.UserId != userClaims.UserId &&
		!userClaims.HasRole(constant.RoleSuperAdmin) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	payment.Auction = &auction
	return payment
}

func (u *paymentUseCase) OwnFetch(ctx context.Context, request dto_request.OwnPaymentFetchRequest) ([]model.Payment, int64) {
	userClaims := model.MustGetUserCtx(ctx)

	option := model.PaymentQueryOption{
		QueryOption: model.NewQueryOptionWithPagination(
			request.Page,
			request.Limit,
			model.Sorts(request.Sorts),
		),
		UserId: util.Pointer(userClaims.UserId),
		Status: request.Status,
	}

	total, err := u.repositoryManager.PaymentRepository().Count(ctx, option)
	panicIfErr(err)

	payments, err := u.repositoryManager.PaymentRepository().Fetch(ctx, option)
	panicIfErr(err)

	return payments, total
}

func (u *paymentUseCase) OwnGet(ctx context.Context, request dto_request.OwnPaymentGetRequest) model.Payment {
	userClaims := model.MustGetUserCtx(ctx)

	payment, err := u.repositoryManager.PaymentRepository().GetById(ctx, request.PaymentId)
	panicIfRepositoryError(err, constant.LanguagePaymentNotFound)

	if payment.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	auction := mustGetAuction(ctx, u.repositoryManager, payment.AuctionId)
	payment.Auction = &auction

	return *payment
}

// createPaymentForWinner creates a Midtrans Snap transaction for an auction
// winner and enqueues a safety-net expiry task.
func (u *paymentUseCase) createPaymentForWinner(
	ctx context.Context,
	auction model.Auction,
	winner model.AuctionWinner,
	bid model.AuctionBid,
	buyer model.User,
) (*model.Payment, error) {
	amount := bid.Amount + auction.Fee

	expiredAt := util.CurrentDateTime().Add(24 * time.Hour)

	// Attach the Midtrans payment method if it exists in the DB.
	var methodId *int64
	if pm, err := u.repositoryManager.PaymentMethodRepository().GetByCode(ctx, constant.PaymentMethodCodeMidtrans); err == nil && pm != nil {
		methodId = &pm.Id
	}

	payment := model.Payment{
		AuctionId:       auction.Id,
		UserId:          buyer.Id,
		PaymentMethodId: methodId,
		Amount:          amount,
		Status:          constant.PaymentStatusWaitingForPayment,
		ExpiredAt:       expiredAt.NullDateTime(),
	}

	if err := u.repositoryManager.PaymentRepository().Insert(ctx, &payment); err != nil {
		return nil, err
	}

	if u.midtransClient != nil {
		paymentId := strconv.FormatInt(payment.Id, 10)
		snapReq := infrastructure.MidtransSnapRequest{
			TransactionDetails: infrastructure.MidtransTransactionDetails{
				OrderId:     paymentId,
				GrossAmount: amount,
			},
			CustomerDetails: infrastructure.MidtransCustomerDetails{
				FirstName: buyer.Fullname,
				Phone:     buyer.Phone,
			},
			Expiry: &infrastructure.MidtransExpiry{
				StartTime: time.Now().Format("2006-01-02 15:04:05 +0700"),
				Unit:      "hour",
				Duration:  24,
			},
		}

		snapResp, err := u.midtransClient.CreateSnapTransaction(snapReq)
		if err == nil && snapResp != nil {
			_, _ = u.repositoryManager.PaymentRepository().UpdateSnapInfo(ctx, payment.Id, snapResp.RedirectUrl, snapResp.Token)
			payment.SnapUrl = &snapResp.RedirectUrl
			payment.SnapToken = &snapResp.Token
		}
	}

	// Enqueue safety-net expiry task (fires 5 minutes after expired_at in case
	// the Midtrans expire webhook is delayed or never arrives).
	if u.taskQueue != nil {
		_ = u.taskQueue.EnqueuePaymentExpiry(payment.Id, expiredAt.Add(5*time.Minute).Time())
	}

	_ = winner
	return &payment, nil
}

// HandleMidtransNotification processes an incoming Midtrans webhook notification.
func (u *paymentUseCase) HandleMidtransNotification(ctx context.Context, notification infrastructure.MidtransNotification) {
	paymentId, parseErr := strconv.ParseInt(notification.OrderId, 10, 64)
	if parseErr != nil {
		return
	}
	payment, err := u.repositoryManager.PaymentRepository().GetById(ctx, paymentId)
	if err != nil {
		// Unknown order – ignore
		return
	}

	// Map Midtrans transaction_status to our internal action.
	//
	// "cancel" / "deny": the Snap session may still be usable (e.g. credit-card
	//   deny lets the buyer retry with another method). We leave the payment as
	//   WAITING_FOR_PAYMENT so a subsequent settlement notification is processed.
	//
	// "expire": the 24-hour window is truly over — trigger the seller-decision flow.
	//
	// "capture" / "settlement": payment confirmed.
	switch notification.TransactionStatus {
	case "capture", "settlement":
		if notification.FraudStatus != "accept" && notification.FraudStatus != "" {
			return // fraud — leave as-is; Midtrans may send a separate deny/cancel
		}
	case "expire":
		// handled below
	default:
		// cancel, deny, pending, and anything else — no action needed
		return
	}

	if payment.Status != constant.PaymentStatusWaitingForPayment {
		return
	}

	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		if notification.TransactionStatus == "expire" {
			updatedPayment, err := u.repositoryManager.PaymentRepository().UpdateStatus(ctx, payment.Id, constant.PaymentStatusExpired)
			if err != nil {
				return err
			}
			return u.onPaymentExpired(ctx, *updatedPayment)
		}

		// settlement / capture
		updatedPayment, err := u.repositoryManager.PaymentRepository().UpdateStatus(ctx, payment.Id, constant.PaymentStatusCompleted)
		if err != nil {
			return err
		}
		return u.onPaymentCompleted(ctx, *updatedPayment)
	}))
}

// onPaymentCompleted marks the winner as COMPLETED and creates a shipment.
func (u *paymentUseCase) onPaymentCompleted(ctx context.Context, payment model.Payment) error {
	// Find and lock the active winner for this auction
	winner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, payment.AuctionId)
	if err != nil {
		return err
	}

	if _, err := u.repositoryManager.AuctionWinnerRepository().UpdateStatus(ctx, winner.Id, constant.AuctionWinnerStatusCompleted); err != nil {
		return err
	}

	if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, payment.AuctionId, constant.AuctionStatusWaitingForBuyerAddress); err != nil {
		return err
	}

	// Resolve buyer's and seller's default addresses (best-effort, may be nil)
	var buyerAddressId *int64
	var sellerAddressId *int64
	var buyerSnapshotStr *string
	var sellerSnapshotStr *string

	buyerAddr, buyerAddrErr := u.repositoryManager.UserAddressRepository().GetDefaultByUserId(ctx, payment.UserId)
	if buyerAddrErr == nil && buyerAddr != nil {
		buyerAddressId = &buyerAddr.Id
		snap := buildAddressSnapshot(buyerAddr)
		buyerSnapshotStr = &snap
	}

	// Resolve seller: look up the auction product to find the seller user_id
	auction, err := u.repositoryManager.AuctionRepository().GetById(ctx, payment.AuctionId)
	if err != nil {
		return err
	}
	product, err := u.repositoryManager.ProductRepository().GetById(ctx, auction.ProductId)
	if err != nil {
		return err
	}
	if _, err := u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusWaitingForBuyerAddress); err != nil {
		return err
	}
	sellerAddr, sellerAddrErr := u.repositoryManager.UserAddressRepository().GetDefaultByUserId(ctx, product.UserId)
	if sellerAddrErr == nil && sellerAddr != nil {
		sellerAddressId = &sellerAddr.Id
		snap := buildAddressSnapshot(sellerAddr)
		sellerSnapshotStr = &snap
	}

	// Create shipment for the winning bid
	shipment := model.Shipment{
		AuctionBidId:          *winner.AuctionBidId,
		UserId:                payment.UserId,
		BuyerAddressId:        buyerAddressId,
		SellerAddressId:       sellerAddressId,
		BuyerAddressSnapshot:  buyerSnapshotStr,
		SellerAddressSnapshot: sellerSnapshotStr,
	}
	if err := u.repositoryManager.ShipmentRepository().Insert(ctx, &shipment); err != nil {
		return err
	}

	// Best-effort: compute Biteship estimated costs and store on shipment
	if u.biteshipClient != nil && buyerAddr != nil && sellerAddr != nil &&
		buyerAddr.BiteshipAreaId != "" && sellerAddr.BiteshipAreaId != "" {
		winnerBidId := *winner.AuctionBidId
		shipmentId := shipment.Id
		biteshipClient := u.biteshipClient
		repoMgr := u.repositoryManager
		go func() {
			bgCtx := context.Background()
			bid, err := repoMgr.AuctionBidRepository().GetById(bgCtx, winnerBidId)
			if err != nil {
				return
			}
			estimated := computeEstimatedCosts(biteshipClient, sellerAddr.BiteshipAreaId, buyerAddr.BiteshipAreaId, product.WeightGram, int(bid.Amount))
			if estimated != "" {
				if _, err := repoMgr.ShipmentRepository().UpdateEstimatedCosts(bgCtx, shipmentId, estimated); err != nil {
					log.Printf("onPaymentCompleted: UpdateEstimatedCosts: %v", err)
				}
			}
		}()
	}

	return nil
}

// onPaymentExpired is called (inside a transaction) when the 24-hour payment
// window elapses without a successful payment.  It cancels the current winner
// and puts the auction into WAITING_FOR_SELLER_DECISION so the seller can
// choose to relist the product or offer it to the next-highest bidder.
func (u *paymentUseCase) onPaymentExpired(ctx context.Context, payment model.Payment) error {
	// Lock and fetch the active winner associated with this auction.
	currentWinner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, payment.AuctionId)
	if err != nil {
		return err
	}

	// Cancel the winner.
	if _, err := u.repositoryManager.AuctionWinnerRepository().UpdateStatus(ctx, currentWinner.Id, constant.AuctionWinnerStatusCancelled); err != nil {
		return err
	}

	// Move the auction to WAITING_FOR_SELLER_DECISION.
	if _, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, payment.AuctionId, constant.AuctionStatusWaitingForSellerDecision); err != nil {
		return err
	}
	// Advance product status to WAITING_FOR_SELLER_DECISION
	auctionForProduct, err := u.repositoryManager.AuctionRepository().GetById(ctx, payment.AuctionId)
	if err != nil {
		return err
	}
	_, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auctionForProduct.ProductId, constant.ProductStatusWaitingForSellerDecision)
	return err
}

// HandlePaymentExpiry is the safety-net asynq task handler.  It fires
// 5 minutes after the payment's expired_at to ensure the expired-payment flow
// runs even when the Midtrans expire webhook is not delivered.
func (u *paymentUseCase) HandlePaymentExpiry(ctx context.Context, paymentId int64) error {
	payment, err := u.repositoryManager.PaymentRepository().GetById(ctx, paymentId)
	if err != nil {
		return err
	}

	// Idempotent: if the webhook already processed this payment, nothing to do.
	if payment.Status != constant.PaymentStatusWaitingForPayment {
		return nil
	}

	return u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		updatedPayment, err := u.repositoryManager.PaymentRepository().UpdateStatus(ctx, payment.Id, constant.PaymentStatusExpired)
		if err != nil {
			return err
		}
		return u.onPaymentExpired(ctx, *updatedPayment)
	})
}

// RecoverExpiredPayments is called on startup to process any payments whose
// expired_at has already passed but are still WAITING_FOR_PAYMENT (happens when
// the server was down when the asynq task should have fired AND Midtrans never
// delivered the expire webhook).
func (u *paymentUseCase) RecoverExpiredPayments(ctx context.Context) error {
	payments, err := u.repositoryManager.PaymentRepository().FetchExpiredWaiting(ctx)
	if err != nil {
		return err
	}
	for _, p := range payments {
		log.Printf("[startup] recovering expired payment %d", p.Id)
		if err := u.HandlePaymentExpiry(ctx, p.Id); err != nil {
			log.Printf("[startup] recover expired payment %d failed: %v", p.Id, err)
		}
	}
	return nil
}

// CreateInitialPaymentForWinner creates the first WAITING_FOR_PAYMENT record
// after an auction closes.  Must be called after HandleCloseAuction succeeds.
// It is idempotent: if the winner already has a payment it returns nil.
func (u *paymentUseCase) CreateInitialPaymentForWinner(ctx context.Context, auctionId int64) error {
	auction, err := u.repositoryManager.AuctionRepository().GetById(ctx, auctionId)
	if err != nil {
		return err
	}

	// Guard: only act when the auction is in WAITING_FOR_PAYMENT state.
	if auction.Status != constant.AuctionStatusWaitingForPayment {
		return nil
	}

	// Get the active winner (ON_GOING status set by closeAuction).
	winner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, auctionId)
	if err == constant.ErrNoData {
		return nil
	}
	if err != nil {
		return err
	}

	// Already has a payment — idempotent.
	if winner.Status == constant.AuctionWinnerStatusWaitingForPayment {
		return nil
	}

	if winner.AuctionBidId == nil {
		log.Printf("[payment] cancel auction %d: active winner %d has no bid", auctionId, winner.Id)
		return u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
			if _, err := u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusCancelled); err != nil {
				return err
			}
			if _, err := u.repositoryManager.AuctionWinnerRepository().UpdateStatus(ctx, winner.Id, constant.AuctionWinnerStatusCancelled); err != nil {
				return err
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
		})
	}

	bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, *winner.AuctionBidId)
	if err != nil {
		return err
	}

	buyer, err := u.repositoryManager.UserRepository().GetById(ctx, bid.UserId)
	if err != nil {
		return err
	}

	return u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		if _, err := u.repositoryManager.AuctionWinnerRepository().UpdateStatus(ctx, winner.Id, constant.AuctionWinnerStatusWaitingForPayment); err != nil {
			return err
		}
		winner.Status = constant.AuctionWinnerStatusWaitingForPayment
		_, err := u.createPaymentForWinner(ctx, *auction, *winner, *bid, *buyer)
		return err
	})
}

// buildAddressSnapshot converts a UserAddress into a JSON string for snapshotting.
func buildAddressSnapshot(a *model.UserAddress) string {
	snap := model.ShipmentAddressSnapshot{
		RecipientName:  a.RecipientName,
		Phone:          a.Phone,
		CityId:         a.CityId,
		CityName:       a.CityName,
		ProvinceName:   a.ProvinceName,
		Address:        a.Address,
		PostalCode:     a.PostalCode,
		BiteshipAreaId: a.BiteshipAreaId,
		Latitude:       a.Latitude,
		Longitude:      a.Longitude,
	}
	b, _ := json.Marshal(snap)
	return string(b)
}

// computeEstimatedCosts fetches shipping options from Biteship and returns a JSON
// string of ShipmentCostEstimate values.  Returns "" on any error.
func computeEstimatedCosts(client infrastructure.BiteshipClient, originAreaId, destAreaId string, weightGram, itemValue int) string {
	options, err := client.Calculate(originAreaId, destAreaId, weightGram, itemValue)
	if err != nil {
		log.Printf("computeEstimatedCosts: %v", err)
		return ""
	}
	var estimates []model.ShipmentCostEstimate
	for _, opt := range options {
		estimates = append(estimates, model.ShipmentCostEstimate{
			CourierName:        opt.CourierName,
			CourierCode:        opt.CourierCode,
			CourierServiceName: opt.CourierServiceName,
			CourierServiceCode: opt.CourierServiceCode,
			ShippingFee:        opt.ShippingFee,
			Price:              opt.Price,
			Duration:           opt.Duration,
		})
	}
	if len(estimates) == 0 {
		return ""
	}
	b, _ := json.Marshal(estimates)
	return string(b)
}
