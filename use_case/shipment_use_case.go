package use_case

import (
	"context"
	"log"
	"strings"

	"auction-service/constant"
	"auction-service/delivery/dto_request"
	"auction-service/delivery/dto_response"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/repository"
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
}

type shipmentUseCase struct {
	repositoryManager repository.RepositoryManager
	biteshipClient    infrastructure.BiteshipClient
}

func NewShipmentUseCase(repositoryManager repository.RepositoryManager, biteshipClient infrastructure.BiteshipClient) ShipmentUseCase {
	return &shipmentUseCase{
		repositoryManager: repositoryManager,
		biteshipClient:    biteshipClient,
	}
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

	var updated *model.Shipment
	panicIfErr(u.repositoryManager.Transaction(ctx, func(ctx context.Context) error {
		var err error
		updated, err = u.repositoryManager.ShipmentRepository().UpdateReceived(ctx, shipment.Id, request.DeliveryProofImagePath)
		if err != nil {
			return err
		}

		bid, err := u.repositoryManager.AuctionBidRepository().GetById(ctx, shipment.AuctionBidId)
		if err != nil {
			return err
		}

		product, err := u.repositoryManager.ProductRepository().GetById(ctx, auction.ProductId)
		if err != nil {
			return err
		}

		if _, err = u.repositoryManager.UserRepository().DepositBalance(ctx, product.UserId, bid.Amount); err != nil {
			return err
		}

		_, err = u.repositoryManager.AuctionRepository().UpdateStatus(ctx, auction.Id, constant.AuctionStatusCompleted)
		if err != nil {
			return err
		}

		_, err = u.repositoryManager.ProductRepository().UpdateStatus(ctx, auction.ProductId, constant.ProductStatusCompleted)
		return err
	}))

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
