package dto_response

import (
	"context"
	"encoding/json"

	"auction-service/data_type"
	"auction-service/model"
)

// ShipmentResponse represents a single shipment in API responses.
type ShipmentResponse struct {
	Id                     string                         `json:"id"                          example:"550e8400-e29b-41d4-a716-446655440000"`
	AuctionBidId           string                         `json:"auction_bid_id"              example:"550e8400-e29b-41d4-a716-446655440001"`
	UserId                 string                         `json:"user_id"                     example:"550e8400-e29b-41d4-a716-446655440002"`
	BuyerAddressId         *string                        `json:"buyer_address_id,omitempty"`
	SellerAddressId        *string                        `json:"seller_address_id,omitempty"`
	BuyerAddressSnapshot   *model.ShipmentAddressSnapshot `json:"buyer_address_snapshot,omitempty"`
	SellerAddressSnapshot  *model.ShipmentAddressSnapshot `json:"seller_address_snapshot,omitempty"`
	ServiceCode            *string                        `json:"service_code,omitempty"`
	ShippingCost           *float64                       `json:"shipping_cost,omitempty"`
	EstimatedCosts         []model.ShipmentCostEstimate   `json:"estimated_costs,omitempty"`
	BiteshipOrderId        *string                        `json:"biteship_order_id,omitempty"`
	TrackingNumber         *string                        `json:"tracking_number,omitempty"`
	CourierCode            *string                        `json:"courier_code,omitempty"`
	DeliveryProofImagePath *string                        `json:"delivery_proof_image_path,omitempty"`
	ShippedAt              data_type.NullDateTime         `json:"shipped_at"`
	ReceivedAt             data_type.NullDateTime         `json:"received_at"`
	AuctionBid             *AuctionBidResponse            `json:"auction_bid,omitempty"`
	Timestamp
} // @name ShipmentResponse

func NewShipmentResponse(ctx context.Context, s model.Shipment) ShipmentResponse {
	r := ShipmentResponse{
		Id:                     s.Id,
		AuctionBidId:           s.AuctionBidId,
		UserId:                 s.UserId,
		BuyerAddressId:         s.BuyerAddressId,
		SellerAddressId:        s.SellerAddressId,
		BuyerAddressSnapshot:   s.ParseBuyerAddressSnapshot(),
		SellerAddressSnapshot:  s.ParseSellerAddressSnapshot(),
		ServiceCode:            s.ServiceCode,
		ShippingCost:           s.ShippingCost,
		BiteshipOrderId:        s.BiteshipOrderId,
		TrackingNumber:         s.TrackingNumber,
		CourierCode:            s.CourierCode,
		DeliveryProofImagePath: s.DeliveryProofImagePath,
		ShippedAt:              s.ShippedAt,
		ReceivedAt:             s.ReceivedAt,
		Timestamp:              Timestamp(s.Timestamp),
	}

	if s.EstimatedCosts != nil && *s.EstimatedCosts != "" {
		var estimates []model.ShipmentCostEstimate
		if err := json.Unmarshal([]byte(*s.EstimatedCosts), &estimates); err == nil {
			r.EstimatedCosts = estimates
		}
	}

	if s.AuctionBid != nil {
		b := NewAuctionBidResponse(ctx, *s.AuctionBid)
		r.AuctionBid = &b
	}

	return r
}
