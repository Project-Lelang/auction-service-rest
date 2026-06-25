package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
)

// PaymentResponse represents a single payment in API responses.
type PaymentResponse struct {
	Id              int64                  `json:"id"               example:"1"`
	AuctionId       int64                  `json:"auction_id"       example:"2"`
	UserId          int64                  `json:"user_id"          example:"3"`
	PaymentMethodId *int64                 `json:"payment_method_id,omitempty"`
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
