package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
)

type SecondChanceOfferResponse struct {
	Id        int64                  `json:"id"         example:"1"`
	AuctionId int64                  `json:"auction_id" example:"2"`
	SellerId  int64                  `json:"seller_id"  example:"3"`
	BuyerId   int64                  `json:"buyer_id"   example:"4"`
	BidId     int64                  `json:"bid_id"     example:"5"`
	Status    string                 `json:"status"     example:"PENDING"`
	ExpiredAt data_type.NullDateTime `json:"expired_at"`
	Auction   *AuctionResponse       `json:"auction,omitempty"`
	Bid       *AuctionBidResponse    `json:"bid,omitempty"`
	Timestamp
} // @name SecondChanceOfferResponse

func NewSecondChanceOfferResponse(ctx context.Context, o model.SecondChanceOffer) SecondChanceOfferResponse {
	r := SecondChanceOfferResponse{
		Id:        o.Id,
		AuctionId: o.AuctionId,
		SellerId:  o.SellerId,
		BuyerId:   o.BuyerId,
		BidId:     o.BidId,
		Status:    o.Status,
		ExpiredAt: o.ExpiredAt,
		Timestamp: Timestamp(o.Timestamp),
	}

	if o.Auction != nil {
		a := NewAuctionResponse(ctx, *o.Auction)
		r.Auction = &a
	}
	if o.Bid != nil {
		b := NewAuctionBidResponse(ctx, *o.Bid)
		r.Bid = &b
	}

	return r
}
