package dto_response

import (
	"context"

	"auction-service/model"
	"auction-service/util"
)

// AuctionBidResponse represents a single auction bid in API responses.
type AuctionBidResponse struct {
	Id        string           `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	UserId    string           `json:"user_id"    example:"550e8400-e29b-41d4-a716-446655440001"`
	AuctionId string           `json:"auction_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Amount    float64          `json:"amount"     example:"150000"`
	IsWinner  bool             `json:"is_winner"  example:"true"`
	User      *UserResponse    `json:"user,omitempty"`
	Auction   *AuctionResponse `json:"auction,omitempty"`
	Payment   *PaymentResponse `json:"payment,omitempty"`
	Timestamp
} // @name AuctionBidResponse

func NewAuctionBidResponse(ctx context.Context, b model.AuctionBid) AuctionBidResponse {
	isWinner := false
	if b.Winner != nil {
		isWinner = true
	}

	r := AuctionBidResponse{
		Id:        b.Id,
		UserId:    b.UserId,
		AuctionId: b.AuctionId,
		Amount:    b.Amount,
		IsWinner:  isWinner,
		Timestamp: Timestamp(b.Timestamp),
	}

	if b.User != nil {
		r.User = util.Pointer(NewUserResponse(ctx, *b.User))
	}

	if b.Auction != nil {
		r.Auction = util.Pointer(NewAuctionResponse(ctx, *b.Auction))
	}

	if b.Payment != nil {
		r.Payment = util.Pointer(NewPaymentResponse(ctx, *b.Payment))
	}

	return r
}
