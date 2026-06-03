package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
)

// PaymentResponse represents a single payment in API responses.
type PaymentResponse struct {
	Id              string                 `json:"id"               example:"550e8400-e29b-41d4-a716-446655440000"`
	AuctionId       string                 `json:"auction_id"       example:"550e8400-e29b-41d4-a716-446655440001"`
	UserId          string                 `json:"user_id"          example:"550e8400-e29b-41d4-a716-446655440002"`
	PaymentMethodId *string                `json:"payment_method_id,omitempty"`
	Amount          float64                `json:"amount"           example:"155000"`
	Status          string                 `json:"status"           example:"WAITING_FOR_PAYMENT"`
	SnapUrl         *string                `json:"snap_url,omitempty"`
	SnapToken       *string                `json:"snap_token,omitempty"`
	ExpiredAt       data_type.NullDateTime `json:"expired_at"`
	Auction         *AuctionResponse       `json:"auction,omitempty"`
	Timestamp
} // @name PaymentResponse

func NewPaymentResponse(ctx context.Context, p model.Payment) PaymentResponse {
	r := PaymentResponse{
		Id:              p.Id,
		AuctionId:       p.AuctionId,
		UserId:          p.UserId,
		PaymentMethodId: p.PaymentMethodId,
		Amount:          p.Amount,
		Status:          p.Status,
		SnapUrl:         p.SnapUrl,
		SnapToken:       p.SnapToken,
		ExpiredAt:       p.ExpiredAt,
		Timestamp:       Timestamp(p.Timestamp),
	}

	if p.Auction != nil {
		a := NewAuctionResponse(ctx, *p.Auction)
		r.Auction = &a
	}

	return r
}
