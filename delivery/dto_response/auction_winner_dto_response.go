package dto_response

import (
	"context"

	"auction-service/model"
)

// AuctionWinnerResponse represents a single auction winner in API responses.
type AuctionWinnerResponse struct {
	Id           int64               `json:"id"             example:"1"`
	AuctionId    int64               `json:"auction_id"     example:"2"`
	AuctionBidId *int64              `json:"auction_bid_id" example:"3"`
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
