package use_case

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/global"
	"auction-service/infrastructure"
	"auction-service/internal/notification"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

// ShipmentUseCase manages shipment lifecycle operations.
type ShipmentUseCase interface {
	FetchByAuction(ctx context.Context, request dto_request.AuctionShipmentFetchRequest) []model.Shipment
	GetByAuction(ctx context.Context, request dto_request.AuctionShipmentGetRequest) model.Shipment
	Ship(ctx context.Context, request dto_request.AuctionShipmentShipRequest) model.Shipment
	Receive(ctx context.Context, request dto_request.AuctionShipmentReceiveRequest) model.Shipment
	UpdateBuyerAddress(ctx context.Context, request dto_request.AuctionShipmentUpdateAddressRequest) model.Shipment
	UpdateSellerAddress(ctx context.Context, request dto_request.AuctionShipmentUpdateAddressRequest) model.Shipment
	GetTracking(ctx context.Context, request dto_request.AuctionShipmentGetTrackingRequest) map[string]interface{}
	HandleShipmentAddressDeadline(ctx context.Context, shipmentId int64) error
	HandleShipmentShipDeadline(ctx context.Context, shipmentId int64) error
	HandleShipmentTrackingCheck(ctx context.Context, shipmentId int64) error
	HandleShipmentReceiveDeadline(ctx context.Context, shipmentId int64) error
	HandleBiteshipWebhook(ctx context.Context, payload map[string]interface{}) error
	RecoverShipmentDeadlines(ctx context.Context) error
}

type shipmentUseCase struct {
	repositoryManager repository.RepositoryManager
	biteshipClient    infrastructure.BiteshipClient
	taskQueue         TaskQueue
	notificationQueue NotificationPublisher
}

func NewShipmentUseCase(repositoryManager repository.RepositoryManager, biteshipClient infrastructure.BiteshipClient, taskQueue TaskQueue, notificationQueue NotificationPublisher) ShipmentUseCase {
	return &shipmentUseCase{
		repositoryManager: repositoryManager,
		biteshipClient:    biteshipClient,
		taskQueue:         taskQueue,
		notificationQueue: notificationQueue,
	}
}

func buyerAddressDuration() time.Duration {
	return time.Duration(global.GetConfig().ShipmentDeadline.BuyerAddressHours) * time.Hour
}

func sellerShipDuration() time.Duration {
	return time.Duration(global.GetConfig().ShipmentDeadline.SellerShipHours) * time.Hour
}

func buyerReceiveDuration() time.Duration {
	return time.Duration(global.GetConfig().ShipmentDeadline.BuyerReceiveHours) * time.Hour
}

func trackingCheckDuration() time.Duration {
	return time.Duration(global.GetConfig().ShipmentDeadline.TrackingCheckIntervalMins) * time.Minute
}

func shipmentDeadlineGrace() time.Duration {
	return time.Duration(global.GetConfig().ShipmentDeadline.DeadlineGraceMinutes) * time.Minute
}

func (u *shipmentUseCase) mustGetAuctionShipment(ctx context.Context, auctionId int64, shipmentId int64) (model.Auction, model.Shipment) {
	auction := mustGetAuction(ctx, u.repositoryManager, auctionId)
	shipment := mustGetShipment(ctx, u.repositoryManager, shipmentId)

	// Verify the shipment belongs to a bid in this auction
	bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, shipment.AuctionBidId)
	panicIfRepositoryError(err, constant.LanguageBidNotFound)
	if bid.AuctionId != auction.Id {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentNotFound))
	}

	return auction, shipment
}

func (u *shipmentUseCase) FetchByAuction(ctx context.Context, request dto_request.AuctionShipmentFetchRequest) []model.Shipment {
	// userClaims := model.MustGetUserCtx(ctx)

	auction := mustGetAuction(ctx, u.repositoryManager, request.AuctionId)

	// Only the seller or superadmin may list all shipments; buyers use GetShipment
	// product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	// if product.UserId != userClaims.UserId && !userClaims.HasRole(constant.RoleSuperAdmin) {
	// 	panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	// }

	// Collect all winning bid IDs for this auction
	winners, err := u.repositoryManager.AuctionWinnerRepository().Fetch(ctx, model.AuctionWinnerQueryOption{
		AuctionId: &auction.Id,
	})
	panicIfErr(err)

	var shipments []model.Shipment
	for _, w := range winners {
		if w.AuctionBidId == nil {
			continue
		}
		s, err := u.repositoryManager.ShipmentRepository().GetByAuctionBidId(ctx, *w.AuctionBidId)
		if err == constant.ErrNoData {
			continue
		}
		panicIfErr(err)
		shipments = append(shipments, *s)
	}

	return shipments
}

func (u *shipmentUseCase) GetByAuction(ctx context.Context, request dto_request.AuctionShipmentGetRequest) model.Shipment {
	userClaims := model.MustGetUserCtx(ctx)

	auction, shipment := u.mustGetAuctionShipment(ctx, request.AuctionId, request.ShipmentId)

	// Only the seller, the buyer, or superadmin may view a shipment
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId &&
		shipment.UserId != userClaims.UserId &&
		!userClaims.HasRole(constant.RoleSuperAdmin) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	return shipment
}

// Ship creates a Komship delivery order, requests pickup, and marks the shipment shipped.
func (u *shipmentUseCase) Ship(ctx context.Context, request dto_request.AuctionShipmentShipRequest) model.Shipment {
	userClaims := model.MustGetUserCtx(ctx)

	auction, shipment := u.mustGetAuctionShipment(ctx, request.AuctionId, request.ShipmentId)

	// Only the auction's product owner (seller) may mark as shipped
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// Buyer must have confirmed their address first
	if auction.Status != constant.AuctionStatusWaitingForShipment {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionAddressNotConfirmed))
	}

	if !shipment.ShippedAt.IsNil() {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentAlreadyShipped))
	}

	if u.biteshipClient == nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemForbidden))
	}

	sellerSnap := shipment.ParseSellerAddressSnapshot()
	buyerSnap := shipment.ParseBuyerAddressSnapshot()
	if sellerSnap == nil || buyerSnap == nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentAddressLocked))
	}

	// Find the winning bid to get item value
	bid := mustGetAuctionBid(ctx, u.repositoryManager, shipment.AuctionBidId)

	// Recalculate rates via Biteship to find the chosen service price.
	// If the rates API fails (e.g. insufficient balance in sandbox), we still
	// proceed with shippingFee = 0 so the shipment can be created.
	options, calcErr := u.biteshipClient.Calculate(
		sellerSnap.BiteshipAreaId,
		buyerSnap.BiteshipAreaId,
		product.WeightGram,
		int(bid.Amount),
	)
	if calcErr != nil {
		log.Printf("Ship: biteship Calculate failed (proceeding with fee=0): %v", calcErr)
	}

	var shippingFee int
	for _, opt := range options {
		if opt.CourierCode == strings.ToLower(request.CourierCode) && opt.CourierServiceCode == strings.ToLower(request.ServiceCode) {
			shippingFee = opt.Price
			break
		}
	}

	// Fetch seller user for contact info
	seller := mustGetUser(ctx, u.repositoryManager, product.UserId)

	orderResult, err := u.biteshipClient.CreateOrder(infrastructure.BiteshipCreateOrderRequest{
		OriginContactName:       seller.Fullname,
		OriginContactPhone:      sellerSnap.Phone,
		OriginAddress:           sellerSnap.Address,
		OriginAreaId:            sellerSnap.BiteshipAreaId,
		DestinationContactName:  buyerSnap.RecipientName,
		DestinationContactPhone: buyerSnap.Phone,
		DestinationAddress:      buyerSnap.Address,
		DestinationAreaId:       buyerSnap.BiteshipAreaId,
		CourierCompany:          strings.ToLower(request.CourierCode),
		CourierType:             strings.ToLower(request.ServiceCode),
		DeliveryType:            "now",
		Items: []infrastructure.BiteshipOrderItem{
			{
				Name:     product.Name,
				Value:    int(bid.Amount),
				Weight:   product.WeightGram,
				Quantity: 1,
				Length:   10,
				Width:    10,
				Height:   10,
			},
		},
	})
	panicIfErr(err)

	// Store the Biteship tracking_id as TrackingNumber (used for GetTracking);
	// waybill_id (AWB) stored in BiteshipOrderId for display purposes.
	updated, err := u.repositoryManager.ShipmentRepository().UpdateShipped(
		ctx,
		shipment.Id,
		orderResult.TrackingId,
		request.CourierCode,
		request.ServiceCode,
		float64(shippingFee),
		orderResult.WaybillId,
	)
	panicIfErr(err)

	// Advance auction status to SHIPPED
	_, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusShipped)
	panicIfErr(err)

	// Advance product status to SHIPPED
	_, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusShipped)
	panicIfErr(err)

	if u.taskQueue != nil {
		if err := u.taskQueue.EnqueueShipmentTrackCheck(updated.Id, time.Now().Add(trackingCheckDuration())); err != nil {
			log.Printf("[shipment worker] enqueue tracking check for %d failed: %v", updated.Id, err)
		}
	}

	return *updated
}

// Receive marks the shipment as received by the buyer.
func (u *shipmentUseCase) Receive(ctx context.Context, request dto_request.AuctionShipmentReceiveRequest) model.Shipment {
	userClaims := model.MustGetUserCtx(ctx)

	auction, shipment := u.mustGetAuctionShipment(ctx, request.AuctionId, request.ShipmentId)

	// Only the buyer may mark as received
	if shipment.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	if shipment.ShippedAt.IsNil() {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentNotShippedYet))
	}
	if !shipment.ReceivedAt.IsNil() {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentAlreadyReceived))
	}

	updated, err := u.completeShipment(ctx, auction, shipment, &request.DeliveryProofImagePath, false)
	panicIfErr(err)

	return *updated
}

// UpdateBuyerAddress allows the buyer to confirm/select their shipping address.
// This is a required step after payment: the auction must be in WAITING_FOR_BUYER_ADDRESS.
// Calling this endpoint confirms the address and advances the auction to WAITING_FOR_SHIPMENT,
// allowing the seller to proceed with shipping.
func (u *shipmentUseCase) UpdateBuyerAddress(ctx context.Context, request dto_request.AuctionShipmentUpdateAddressRequest) model.Shipment {
	userClaims := model.MustGetUserCtx(ctx)

	auction, shipment := u.mustGetAuctionShipment(ctx, request.AuctionId, request.ShipmentId)

	if shipment.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// Buyer may only confirm address during the dedicated address-selection window
	if auction.Status != constant.AuctionStatusWaitingForBuyerAddress {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionNotWaitingForBuyerAddress))
	}
	if !shipment.BuyerAddressDeadlineAt.IsNil() && !shipment.BuyerAddressDeadlineAt.DateTime().Time().After(time.Now()) {
		_ = u.HandleShipmentAddressDeadline(ctx, shipment.Id)
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageAuctionNotWaitingForBuyerAddress))
	}

	// Verify the address belongs to the buyer
	address, err := u.repositoryManager.UserAddressRepository().GetById(ctx, request.AddressId)
	panicIfRepositoryError(err, constant.LanguageUserAddressNotFound)
	if address.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageUserAddressNotOwned))
	}

	snapshot := buildAddressSnapshot(address)
	updated, err := u.repositoryManager.ShipmentRepository().UpdateBuyerAddress(ctx, shipment.Id, address.Id, snapshot)
	panicIfErr(err)

	// Address confirmed — unlock seller to proceed with shipping
	_, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusWaitingForShipment)
	panicIfErr(err)

	// Advance product status to WAITING_FOR_SHIPMENT
	_, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusWaitingForShipment)
	panicIfErr(err)

	shipDeadline := util.CurrentDateTime().Add(sellerShipDuration())
	updated, err = u.repositoryManager.ShipmentRepository().UpdateShipDeadline(ctx, shipment.Id, shipDeadline)
	panicIfErr(err)
	if u.taskQueue != nil {
		if err := u.taskQueue.EnqueueShipmentShipDue(updated.Id, shipDeadline.Add(shipmentDeadlineGrace()).Time()); err != nil {
			log.Printf("[shipment worker] enqueue ship deadline for %d failed: %v", updated.Id, err)
		}
	}

	// FOR TESTING PURPOSES: if EstimatedCosts is empty, set it to a non-empty JSON so that the shipping flow can proceed without calling the Biteship rates API.
	if updated.EstimatedCosts == nil || *updated.EstimatedCosts == "" {
		zeroEstimatedCosts := `[
  {
    "price": 12000,
    "duration": "2-3 Hari",
    "courier_code": "jne",
    "courier_name": "JNE Express",
    "shipping_fee": 12000,
    "courier_service_code": "reg",
    "courier_service_name": "Reguler"
  },
  {
    "price": 24000,
    "duration": "1 Hari",
    "courier_code": "jne",
    "courier_name": "JNE Express",
    "shipping_fee": 24000,
    "courier_service_code": "yes",
    "courier_service_name": "Yakin Esok Sampai"
  },
  {
    "price": 11000,
    "duration": "2-4 Hari",
    "courier_code": "jnt",
    "courier_name": "J&T Express",
    "shipping_fee": 11000,
    "courier_service_code": "ez",
    "courier_service_name": "EZ (Regular)"
  },
  {
    "price": 12000,
    "duration": "1-2 Hari",
    "courier_code": "sicepat",
    "courier_name": "SiCepat Ekspres",
    "shipping_fee": 12000,
    "courier_service_code": "siuntung",
    "courier_service_name": "SiUntung"
  },
  {
    "price": 30000,
    "duration": "1-2 Hari",
    "courier_code": "sicepat",
    "courier_name": "SiCepat Ekspres",
    "shipping_fee": 30000,
    "courier_service_code": "best",
    "courier_service_name": "Besok Sampai Tujuan"
  }
]`
		updatedEstimate, err := u.repositoryManager.ShipmentRepository().UpdateEstimatedCosts(ctx, shipment.Id, zeroEstimatedCosts)
		panicIfErr(err)
		updated = updatedEstimate
	}

	return *updated
}

// UpdateSellerAddress allows the seller to set/select their sender address (before shipped).
func (u *shipmentUseCase) UpdateSellerAddress(ctx context.Context, request dto_request.AuctionShipmentUpdateAddressRequest) model.Shipment {
	userClaims := model.MustGetUserCtx(ctx)

	auction, shipment := u.mustGetAuctionShipment(ctx, request.AuctionId, request.ShipmentId)

	// Only the product owner (seller) may set the seller address
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	// Shipment must not be already shipped
	if !shipment.ShippedAt.IsNil() {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentAlreadyShipped))
	}

	// Verify the address belongs to the seller
	address, err := u.repositoryManager.UserAddressRepository().GetById(ctx, request.AddressId)
	panicIfRepositoryError(err, constant.LanguageUserAddressNotFound)
	if address.UserId != userClaims.UserId {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageUserAddressNotOwned))
	}

	snapshot := buildAddressSnapshot(address)
	updated, err := u.repositoryManager.ShipmentRepository().UpdateSellerAddress(ctx, shipment.Id, address.Id, snapshot)
	panicIfErr(err)

	return *updated
}

// GetTracking retrieves live tracking information via Komship.
func (u *shipmentUseCase) GetTracking(ctx context.Context, request dto_request.AuctionShipmentGetTrackingRequest) map[string]interface{} {
	userClaims := model.MustGetUserCtx(ctx)

	auction, shipment := u.mustGetAuctionShipment(ctx, request.AuctionId, request.ShipmentId)

	// Buyer, seller or superadmin may track
	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	if product.UserId != userClaims.UserId &&
		shipment.UserId != userClaims.UserId &&
		!userClaims.HasRole(constant.RoleSuperAdmin) {
		panic(dto_response.NewForbiddenErrorResponse(constant.LanguageSystemForbidden))
	}

	if shipment.TrackingNumber == nil || shipment.CourierCode == nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageShipmentNoTrackingNumber))
	}

	if u.biteshipClient == nil {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemForbidden))
	}

	// TrackingNumber stores the Biteship tracking_id assigned at order creation.
	result, err := u.biteshipClient.GetTracking(*shipment.TrackingNumber)
	panicIfErr(err)

	return result
}

func (u *shipmentUseCase) completeShipment(ctx context.Context, auction model.Auction, shipment model.Shipment, deliveryProofImagePath *string, auto bool) (*model.Shipment, error) {
	var updated *model.Shipment
	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		lockedShipment, err := u.repositoryManager.ShipmentRepository().GetByIdForUpdate(ctx, shipment.Id)
		if err != nil {
			return err
		}
		if !lockedShipment.ReceivedAt.IsNil() {
			updated = lockedShipment
			return nil
		}
		if lockedShipment.ShippedAt.IsNil() || !lockedShipment.SellerFailedAt.IsNil() {
			return nil
		}

		if auto {
			updated, err = u.repositoryManager.ShipmentRepository().UpdateAutoReceived(ctx, lockedShipment.Id)
		} else {
			if deliveryProofImagePath == nil {
				return nil
			}
			updated, err = u.repositoryManager.ShipmentRepository().UpdateReceived(ctx, lockedShipment.Id, *deliveryProofImagePath)
		}
		if err != nil {
			return err
		}

		bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, lockedShipment.AuctionBidId)
		if err != nil {
			return err
		}

		product, err := u.repositoryManager.ProductRepository().GetById(ctx, auction.ProductId)
		if err != nil {
			return err
		}

		sellerRevenue := roundMoney(bid.Amount - calculateAuctionFee(bid.Amount))
		if _, err = u.repositoryManager.UserRepository().DepositBalance(ctx, product.UserId, sellerRevenue); err != nil {
			return err
		}

		if _, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusCompleted); err != nil {
			return err
		}

		_, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusCompleted)
		return err
	})
	return updated, err
}

func (u *shipmentUseCase) refundCompletedPayment(ctx context.Context, auctionId int64) (*model.Payment, error) {
	payment, err := u.repositoryManager.PaymentRepository().GetCompletedByAuctionId(ctx, auctionId)
	if err == constant.ErrNoData {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err = u.repositoryManager.PaymentRepository().UpdateStatus(ctx, payment.Id, constant.PaymentStatusRefunded); err != nil {
		return nil, err
	}
	if _, err = u.repositoryManager.UserRepository().DepositBalance(ctx, payment.UserId, payment.Amount); err != nil {
		return nil, err
	}
	payment.Status = constant.PaymentStatusRefunded
	return payment, nil
}

func (u *shipmentUseCase) HandleShipmentAddressDeadline(ctx context.Context, shipmentId int64) error {
	auction, shipment := u.mustGetAuctionShipmentFromShipmentId(ctx, shipmentId)
	if auction.Status != constant.AuctionStatusWaitingForBuyerAddress ||
		shipment.BuyerAddressDeadlineAt.IsNil() ||
		!shipment.BuyerAddressFailedAt.IsNil() ||
		shipment.BuyerAddressDeadlineAt.DateTime().Time().After(time.Now()) {
		return nil
	}

	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	var payment *model.Payment

	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		lockedShipment, err := u.repositoryManager.ShipmentRepository().GetByIdForUpdate(ctx, shipment.Id)
		if err != nil {
			return err
		}
		if !lockedShipment.BuyerAddressFailedAt.IsNil() {
			return nil
		}
		if _, err = u.repositoryManager.ShipmentRepository().UpdateBuyerAddressFailed(ctx, lockedShipment.Id); err != nil {
			return err
		}
		if _, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusWaitingForSellerDecision); err != nil {
			return err
		}
		if _, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusWaitingForSellerDecision); err != nil {
			return err
		}

		winner, err := u.repositoryManager.AuctionWinnerRepository().GetActiveByAuctionIdForUpdate(ctx, auction.Id)
		if err != nil && err != constant.ErrNoData {
			return err
		}
		if winner != nil {
			if _, err = u.repositoryManager.AuctionWinnerRepository().UpdateStatus(ctx, winner.Id, constant.AuctionWinnerStatusCancelled); err != nil {
				return err
			}
		}

		payment, err = u.refundCompletedPayment(ctx, auction.Id)
		if err != nil {
			return err
		}

		msg := "Buyer did not confirm shipping address before the deadline"
		return u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
			ProductId: auction.ProductId,
			Status:    constant.ProductStatusWaitingForSellerDecision,
			Message:   &msg,
		})
	})
	if err != nil {
		return err
	}

	ref := auction.Id
	insertUserNotification(ctx, u.repositoryManager, shipment.UserId, "Address confirmation expired", "You missed the address confirmation deadline, so your payment was refunded to your balance.", "BUYER_ADDRESS_EXPIRED", &ref)
	insertUserNotification(ctx, u.repositoryManager, product.UserId, "Buyer address deadline missed", "Buyer did not confirm the shipping address before the deadline. You can relist or offer the auction to the next bidder.", "BUYER_ADDRESS_EXPIRED", &ref)
	if payment != nil {
		u.publishShipmentNotification(ctx, shipment.UserId, notification.RoleBuyer, notification.EventBuyerAddressExpired, auction.Id, "Address confirmation expired", "You missed the address confirmation deadline, so your payment was refunded.")
	}
	u.publishShipmentNotification(ctx, product.UserId, notification.RoleSeller, notification.EventBuyerAddressExpired, auction.Id, "Buyer address deadline missed", "Buyer did not confirm the shipping address before the deadline.")
	return nil
}

func (u *shipmentUseCase) HandleShipmentShipDeadline(ctx context.Context, shipmentId int64) error {
	auction, shipment := u.mustGetAuctionShipmentFromShipmentId(ctx, shipmentId)
	if auction.Status != constant.AuctionStatusWaitingForShipment ||
		shipment.ShipDeadlineAt.IsNil() ||
		!shipment.ShippedAt.IsNil() ||
		!shipment.SellerFailedAt.IsNil() ||
		shipment.ShipDeadlineAt.DateTime().Time().After(time.Now()) {
		return nil
	}

	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)

	err := u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		lockedShipment, err := u.repositoryManager.ShipmentRepository().GetByIdForUpdate(ctx, shipment.Id)
		if err != nil {
			return err
		}
		if !lockedShipment.ShippedAt.IsNil() || !lockedShipment.SellerFailedAt.IsNil() {
			return nil
		}

		if _, err = u.repositoryManager.ShipmentRepository().UpdateSellerFailed(ctx, lockedShipment.Id); err != nil {
			return err
		}
		if _, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusSellerFailedToShip); err != nil {
			return err
		}
		if _, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusSellerFailedToShip); err != nil {
			return err
		}

		if _, err = u.refundCompletedPayment(ctx, auction.Id); err != nil {
			return err
		}

		msg := "Seller did not ship before the deadline"
		return u.repositoryManager.ProductStatusHistoryRepository().Insert(ctx, &model.ProductStatusHistory{
			ProductId: auction.ProductId,
			Status:    constant.ProductStatusSellerFailedToShip,
			Message:   &msg,
		})
	})
	if err != nil {
		return err
	}

	ref := auction.Id
	insertUserNotification(ctx, u.repositoryManager, shipment.UserId, "Order refunded", "Seller did not ship before the deadline, so your payment was refunded to your balance.", "SHIPMENT_REFUNDED", &ref)
	insertUserNotification(ctx, u.repositoryManager, product.UserId, "Shipment deadline missed", "You missed the shipping deadline for this auction.", "SHIPMENT_DEADLINE_MISSED", &ref)
	u.publishShipmentNotification(ctx, shipment.UserId, notification.RoleBuyer, notification.EventShipmentRefunded, auction.Id, "Order refunded", "Seller did not ship before the deadline, so your payment was refunded.")
	u.publishShipmentNotification(ctx, product.UserId, notification.RoleSeller, notification.EventShipmentDeadlineMissed, auction.Id, "Shipment deadline missed", "You missed the shipping deadline for this auction.")
	return nil
}

func (u *shipmentUseCase) HandleShipmentTrackingCheck(ctx context.Context, shipmentId int64) error {
	auction, shipment := u.mustGetAuctionShipmentFromShipmentId(ctx, shipmentId)
	if auction.Status != constant.AuctionStatusShipped ||
		shipment.TrackingNumber == nil ||
		shipment.ShippedAt.IsNil() ||
		!shipment.DeliveredAt.IsNil() ||
		!shipment.ReceivedAt.IsNil() ||
		!shipment.SellerFailedAt.IsNil() {
		return nil
	}
	if u.biteshipClient == nil {
		return nil
	}

	tracking, err := u.biteshipClient.GetTracking(*shipment.TrackingNumber)
	if err != nil {
		log.Printf("[shipment worker] tracking check shipment %d failed: %v", shipment.Id, err)
		u.reenqueueTrackingCheck(shipment.Id)
		return nil
	}
	if !isBiteshipDeliveredTracking(tracking) {
		u.reenqueueTrackingCheck(shipment.Id)
		return nil
	}

	_, _ = auction, shipment
	return u.markShipmentDelivered(ctx, shipmentId)
}

func (u *shipmentUseCase) HandleBiteshipWebhook(ctx context.Context, payload map[string]interface{}) error {
	if !isBiteshipDeliveredTracking(payload) {
		return nil
	}
	identifiers := extractBiteshipTrackingIdentifiers(payload)
	shipment, err := u.repositoryManager.ShipmentRepository().GetByTrackingIdentifier(ctx, identifiers)
	if err == constant.ErrNoData {
		log.Printf("[biteship webhook] delivered event ignored: shipment not found identifiers=%v", identifiers)
		return nil
	}
	if err != nil {
		return err
	}
	return u.markShipmentDelivered(ctx, shipment.Id)
}

func (u *shipmentUseCase) HandleShipmentReceiveDeadline(ctx context.Context, shipmentId int64) error {
	auction, shipment := u.mustGetAuctionShipmentFromShipmentId(ctx, shipmentId)
	if auction.Status != constant.AuctionStatusShipped ||
		shipment.ReceiveDeadlineAt.IsNil() ||
		shipment.ShippedAt.IsNil() ||
		!shipment.ReceivedAt.IsNil() ||
		!shipment.SellerFailedAt.IsNil() ||
		shipment.ReceiveDeadlineAt.DateTime().Time().After(time.Now()) {
		return nil
	}

	updated, err := u.completeShipment(ctx, auction, shipment, nil, true)
	if err != nil {
		return err
	}
	if updated == nil {
		return nil
	}

	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	ref := auction.Id
	insertUserNotification(ctx, u.repositoryManager, shipment.UserId, "Order auto-completed", "Receipt confirmation deadline passed, so the order was completed automatically.", "SHIPMENT_AUTO_COMPLETED", &ref)
	insertUserNotification(ctx, u.repositoryManager, product.UserId, "Order completed", "Buyer confirmation deadline passed, so the order was completed automatically.", "SHIPMENT_AUTO_COMPLETED", &ref)
	u.publishShipmentNotification(ctx, shipment.UserId, notification.RoleBuyer, notification.EventShipmentAutoCompleted, auction.Id, "Order auto-completed", "Receipt confirmation deadline passed, so the order was completed automatically.")
	u.publishShipmentNotification(ctx, product.UserId, notification.RoleSeller, notification.EventShipmentAutoCompleted, auction.Id, "Order completed", "Buyer confirmation deadline passed, so the order was completed automatically.")
	return nil
}

func (u *shipmentUseCase) RecoverShipmentDeadlines(ctx context.Context) error {
	shipments, err := u.repositoryManager.ShipmentRepository().FetchPendingAddressDeadline(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, shipment := range shipments {
		if shipment.BuyerAddressDeadlineAt.DateTime().Time().After(now) {
			if u.taskQueue != nil {
				if err := u.taskQueue.EnqueueShipmentAddressDue(shipment.Id, shipment.BuyerAddressDeadlineAt.DateTime().Add(shipmentDeadlineGrace()).Time()); err != nil {
					log.Printf("[startup] re-enqueue address deadline shipment %d failed: %v", shipment.Id, err)
				}
			}
			continue
		}
		log.Printf("[startup] recovering overdue address deadline shipment %d", shipment.Id)
		if err := u.HandleShipmentAddressDeadline(ctx, shipment.Id); err != nil {
			log.Printf("[startup] recover address deadline shipment %d failed: %v", shipment.Id, err)
		}
	}

	shipments, err = u.repositoryManager.ShipmentRepository().FetchPendingShipDeadline(ctx)
	if err != nil {
		return err
	}
	for _, shipment := range shipments {
		if shipment.ShipDeadlineAt.DateTime().Time().After(now) {
			if u.taskQueue != nil {
				if err := u.taskQueue.EnqueueShipmentShipDue(shipment.Id, shipment.ShipDeadlineAt.DateTime().Add(shipmentDeadlineGrace()).Time()); err != nil {
					log.Printf("[startup] re-enqueue ship deadline shipment %d failed: %v", shipment.Id, err)
				}
			}
			continue
		}
		log.Printf("[startup] recovering overdue ship deadline shipment %d", shipment.Id)
		if err := u.HandleShipmentShipDeadline(ctx, shipment.Id); err != nil {
			log.Printf("[startup] recover ship deadline shipment %d failed: %v", shipment.Id, err)
		}
	}

	shipments, err = u.repositoryManager.ShipmentRepository().FetchPendingReceiveDeadline(ctx)
	if err != nil {
		return err
	}
	for _, shipment := range shipments {
		if shipment.ReceiveDeadlineAt.DateTime().Time().After(now) {
			if u.taskQueue != nil {
				if err := u.taskQueue.EnqueueShipmentReceiveDue(shipment.Id, shipment.ReceiveDeadlineAt.DateTime().Add(shipmentDeadlineGrace()).Time()); err != nil {
					log.Printf("[startup] re-enqueue receive deadline shipment %d failed: %v", shipment.Id, err)
				}
			}
			continue
		}
		log.Printf("[startup] recovering overdue receive deadline shipment %d", shipment.Id)
		if err := u.HandleShipmentReceiveDeadline(ctx, shipment.Id); err != nil {
			log.Printf("[startup] recover receive deadline shipment %d failed: %v", shipment.Id, err)
		}
	}

	shipments, err = u.repositoryManager.ShipmentRepository().FetchPendingDeliveryTracking(ctx)
	if err != nil {
		return err
	}
	for _, shipment := range shipments {
		if u.taskQueue != nil {
			if err := u.taskQueue.EnqueueShipmentTrackCheck(shipment.Id, time.Now().Add(trackingCheckDuration())); err != nil {
				log.Printf("[startup] re-enqueue tracking check shipment %d failed: %v", shipment.Id, err)
			}
		}
	}
	return nil
}

func (u *shipmentUseCase) mustGetAuctionShipmentFromShipmentId(ctx context.Context, shipmentId int64) (model.Auction, model.Shipment) {
	shipment := mustGetShipment(ctx, u.repositoryManager, shipmentId)
	bid := mustGetAuctionBid(ctx, u.repositoryManager, shipment.AuctionBidId)
	auction := mustGetAuction(ctx, u.repositoryManager, bid.AuctionId)
	return auction, shipment
}

func (u *shipmentUseCase) markShipmentDelivered(ctx context.Context, shipmentId int64) error {
	auction, shipment := u.mustGetAuctionShipmentFromShipmentId(ctx, shipmentId)
	if auction.Status != constant.AuctionStatusShipped ||
		shipment.ShippedAt.IsNil() ||
		!shipment.DeliveredAt.IsNil() ||
		!shipment.ReceivedAt.IsNil() ||
		!shipment.SellerFailedAt.IsNil() {
		return nil
	}

	receiveDeadline := util.CurrentDateTime().Add(buyerReceiveDuration())
	updated, err := u.repositoryManager.ShipmentRepository().UpdateDelivered(ctx, shipment.Id, receiveDeadline)
	if err != nil {
		return err
	}
	if u.taskQueue != nil {
		if err := u.taskQueue.EnqueueShipmentReceiveDue(updated.Id, receiveDeadline.Add(shipmentDeadlineGrace()).Time()); err != nil {
			log.Printf("[shipment worker] enqueue receive deadline for %d failed: %v", updated.Id, err)
		}
	}

	product := mustGetProduct(ctx, u.repositoryManager, auction.ProductId)
	ref := auction.Id
	insertUserNotification(ctx, u.repositoryManager, shipment.UserId, "Package delivered", "Courier tracking says your package has been delivered. Please confirm receipt before the deadline.", "SHIPMENT_DELIVERED", &ref)
	insertUserNotification(ctx, u.repositoryManager, product.UserId, "Package delivered", "Courier tracking says the package has been delivered. Buyer confirmation deadline has started.", "SHIPMENT_DELIVERED", &ref)
	u.publishShipmentNotification(ctx, shipment.UserId, notification.RoleBuyer, notification.EventShipmentDelivered, auction.Id, "Package delivered", "Courier tracking says your package has been delivered. Please confirm receipt before the deadline.")
	u.publishShipmentNotification(ctx, product.UserId, notification.RoleSeller, notification.EventShipmentDelivered, auction.Id, "Package delivered", "Courier tracking says the package has been delivered. Buyer confirmation deadline has started.")
	return nil
}

func (u *shipmentUseCase) publishShipmentNotification(ctx context.Context, userId int64, role string, eventType string, auctionId int64, title string, body string) {
	publishAuctionNotification(ctx, u.notificationQueue, notification.Payload{
		UserId:    userId,
		Role:      role,
		EventType: eventType,
		AuctionId: auctionId,
		Title:     title,
		Body:      body,
		DataPayload: map[string]string{
			"auction_url": auctionURL(auctionId),
			"auction_id":  strconv.FormatInt(auctionId, 10),
		},
	})
}

func (u *shipmentUseCase) reenqueueTrackingCheck(shipmentId int64) {
	if u.taskQueue == nil {
		return
	}
	if err := u.taskQueue.EnqueueShipmentTrackCheck(shipmentId, time.Now().Add(trackingCheckDuration())); err != nil {
		log.Printf("[shipment worker] re-enqueue tracking check for %d failed: %v", shipmentId, err)
	}
}

func isBiteshipDeliveredTracking(value interface{}) bool {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, raw := range v {
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "status") || strings.Contains(lowerKey, "state") {
				if isDeliveredStatus(raw) {
					return true
				}
			}
			if isBiteshipDeliveredTracking(raw) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if isBiteshipDeliveredTracking(item) {
				return true
			}
		}
	}
	return false
}

func isDeliveredStatus(value interface{}) bool {
	text := strings.ToLower(strings.TrimSpace(toStatusString(value)))
	if text == "" {
		return false
	}
	deliveredStatuses := []string{
		"delivered",
		"received",
		"completed",
		"success",
		"terkirim",
		"diterima",
	}
	for _, status := range deliveredStatuses {
		if text == status || strings.Contains(text, status) {
			return true
		}
	}
	return false
}

func toStatusString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case float64, float32, int, int64, int32, uint, uint64, uint32:
		return fmt.Sprint(v)
	default:
		return ""
	}
}

func extractBiteshipTrackingIdentifiers(value interface{}) []string {
	var out []string
	var walk func(interface{})
	walk = func(v interface{}) {
		switch current := v.(type) {
		case map[string]interface{}:
			for key, raw := range current {
				lowerKey := strings.ToLower(key)
				if strings.Contains(lowerKey, "tracking") ||
					strings.Contains(lowerKey, "waybill") ||
					lowerKey == "awb" ||
					lowerKey == "awb_id" ||
					lowerKey == "resi" {
					if text := toStatusString(raw); strings.TrimSpace(text) != "" {
						out = append(out, strings.TrimSpace(text))
					}
				}
				walk(raw)
			}
		case []interface{}:
			for _, item := range current {
				walk(item)
			}
		}
	}
	walk(value)
	return out
}
