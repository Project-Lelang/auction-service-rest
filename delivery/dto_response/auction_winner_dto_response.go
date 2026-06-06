package dto_response

import (
	"context"

	"auction-service/model"
)

// AuctionWinnerResponse represents a single auction winner in API responses.
type AuctionWinnerResponse struct {
	Id           string              `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000"`
	AuctionId    string              `json:"auction_id"     example:"550e8400-e29b-41d4-a716-446655440001"`
	AuctionBidId string              `json:"auction_bid_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Status       string              `json:"status"         example:"WAITING_FOR_PAYMENT"`
	Auction      *AuctionResponse    `json:"auction,omitempty"`
	AuctionBid   *AuctionBidResponse `json:"auction_bid,omitempty"`
	Timestamp
} // @name AuctionWinnerResponse

func NewAuctionWinnerResponse(ctx context.Context, w model.AuctionWinner) AuctionWinnerResponse {
	r := AuctionWinnerResponse{
		Id:           w.Id,
		AuctionId:    w.AuctionId,
		AuctionBidId: w.AuctionBidId,
		Status:       w.Status,
		Timestamp:    Timestamp(w.Timestamp),
	}

	if w.Auction != nil {
		a := NewAuctionResponse(ctx, *w.Auction)
		r.Auction = &a
	}
	if w.AuctionBid != nil {
		b := NewAuctionBidResponse(ctx, *w.AuctionBid)
		r.AuctionBid = &b
	}

	return r
}
