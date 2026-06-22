package dto_response

import (
	"context"

	"auction-service/model"
	"auction-service/util"
)

// AuctionBidResponse represents a single auction bid in API responses.
type AuctionBidResponse struct {
	Id        int64            `json:"id"         example:"1"`
	UserId    int64            `json:"user_id"    example:"2"`
	AuctionId int64            `json:"auction_id" example:"3"`
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
