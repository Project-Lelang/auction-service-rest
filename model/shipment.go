package model

import (
	"encoding/json"

	"auction-service/data_type"
)

const ShipmentTableName = "shipments"

// ShipmentAddressSnapshot is stored as JSON inside shipments to preserve the
// address at the time of shipment creation, even if the user later edits it.
type ShipmentAddressSnapshot struct {
	RecipientName  string   `json:"recipient_name"`
	Phone          string   `json:"phone"`
	CityId         string   `json:"city_id"`
	CityName       string   `json:"city_name"`
	ProvinceName   string   `json:"province_name"`
	Address        string   `json:"address"`
	PostalCode     string   `json:"postal_code"`
	BiteshipAreaId string   `json:"biteship_area_id"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
}

// ShipmentCostEstimate represents a single courier/service cost option from
// Biteship, stored as part of the estimated_costs JSON array.
type ShipmentCostEstimate struct {
	CourierName        string `json:"courier_name"`
	CourierCode        string `json:"courier_code"`
	CourierServiceName string `json:"courier_service_name"`
	CourierServiceCode string `json:"courier_service_code"`
	ShippingFee        int    `json:"shipping_fee"`
	Price              int    `json:"price"`
	Duration           string `json:"duration"`
}

type Shipment struct {
	Id                     int64                  `db:"id"`
	AuctionBidId           int64                  `db:"auction_bid_id"`
	UserId                 int64                  `db:"user_id"`
	BuyerAddressId         *int64                 `db:"buyer_address_id"`
	SellerAddressId        *int64                 `db:"seller_address_id"`
	BuyerAddressSnapshot   *string                `db:"buyer_address_snapshot"`  // JSON
	SellerAddressSnapshot  *string                `db:"seller_address_snapshot"` // JSON
	ServiceCode            *string                `db:"service_code"`
	ShippingCost           *float64               `db:"shipping_cost"`
	EstimatedCosts         *string                `db:"estimated_costs"` // JSON
	BiteshipOrderId        *string                `db:"biteship_order_id"`
	TrackingNumber         *string                `db:"tracking_number"`
	CourierCode            *string                `db:"courier_code"`
	DeliveryProofImagePath *string                `db:"delivery_proof_image_path"`
	BuyerAddressDeadlineAt data_type.NullDateTime `db:"buyer_address_deadline_at"`
	ShipDeadlineAt         data_type.NullDateTime `db:"ship_deadline_at"`
	ReceiveDeadlineAt      data_type.NullDateTime `db:"receive_deadline_at"`
	DeliveredAt            data_type.NullDateTime `db:"delivered_at"`
	ShippedAt              data_type.NullDateTime `db:"shipped_at"`
	ReceivedAt             data_type.NullDateTime `db:"received_at"`
	AutoReceivedAt         data_type.NullDateTime `db:"auto_received_at"`
	BuyerAddressFailedAt   data_type.NullDateTime `db:"buyer_address_failed_at"`
	SellerFailedAt         data_type.NullDateTime `db:"seller_failed_at"`
	Timestamp

	// relations
	AuctionBid *AuctionBid `db:"-"`
	User       *User       `db:"-"`
}

func (s *Shipment) TableName() string { return ShipmentTableName }

func (s *Shipment) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                        s.Id,
		"auction_bid_id":            s.AuctionBidId,
		"user_id":                   s.UserId,
		"buyer_address_id":          s.BuyerAddressId,
		"seller_address_id":         s.SellerAddressId,
		"buyer_address_snapshot":    s.BuyerAddressSnapshot,
		"seller_address_snapshot":   s.SellerAddressSnapshot,
		"service_code":              s.ServiceCode,
		"shipping_cost":             s.ShippingCost,
		"estimated_costs":           s.EstimatedCosts,
		"biteship_order_id":         s.BiteshipOrderId,
		"tracking_number":           s.TrackingNumber,
		"courier_code":              s.CourierCode,
		"delivery_proof_image_path": s.DeliveryProofImagePath,
		"buyer_address_deadline_at": s.BuyerAddressDeadlineAt,
		"ship_deadline_at":          s.ShipDeadlineAt,
		"receive_deadline_at":       s.ReceiveDeadlineAt,
		"delivered_at":              s.DeliveredAt,
		"shipped_at":                s.ShippedAt,
		"received_at":               s.ReceivedAt,
		"auto_received_at":          s.AutoReceivedAt,
		"buyer_address_failed_at":   s.BuyerAddressFailedAt,
		"seller_failed_at":          s.SellerFailedAt,
		"created_at":                s.CreatedAt,
		"updated_at":                s.UpdatedAt,
	}
}

// ParseEstimatedCosts deserialises the estimated_costs JSON column.
func (s *Shipment) ParseEstimatedCosts() []ShipmentCostEstimate {
	if s.EstimatedCosts == nil || *s.EstimatedCosts == "" {
		return nil
	}
	var out []ShipmentCostEstimate
	_ = json.Unmarshal([]byte(*s.EstimatedCosts), &out)
	return out
}

// ParseBuyerAddressSnapshot deserialises the buyer_address_snapshot JSON column.
func (s *Shipment) ParseBuyerAddressSnapshot() *ShipmentAddressSnapshot {
	if s.BuyerAddressSnapshot == nil {
		return nil
	}
	var out ShipmentAddressSnapshot
	if err := json.Unmarshal([]byte(*s.BuyerAddressSnapshot), &out); err != nil {
		return nil
	}
	return &out
}

// ParseSellerAddressSnapshot deserialises the seller_address_snapshot JSON column.
func (s *Shipment) ParseSellerAddressSnapshot() *ShipmentAddressSnapshot {
	if s.SellerAddressSnapshot == nil {
		return nil
	}
	var out ShipmentAddressSnapshot
	if err := json.Unmarshal([]byte(*s.SellerAddressSnapshot), &out); err != nil {
		return nil
	}
	return &out
}

type ShipmentQueryOption struct {
	QueryOption

	AuctionBidId *int64
	UserId       *int64
}

var _ PrepareOption = &ShipmentQueryOption{}

func (o *ShipmentQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *ShipmentQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"s." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
