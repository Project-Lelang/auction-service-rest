package model

import "auction-service/data_type"

const SecondChanceOfferTableName = "second_chance_offers"

type SecondChanceOffer struct {
	Id        int64                  `db:"id"`
	AuctionId int64                  `db:"auction_id"`
	SellerId  int64                  `db:"seller_id"`
	BuyerId   int64                  `db:"buyer_id"`
	BidId     int64                  `db:"bid_id"`
	Status    string                 `db:"status"`
	ExpiredAt data_type.NullDateTime `db:"expired_at"`
	Timestamp

	Auction *Auction    `db:"-"`
	Bid     *AuctionBid `db:"-"`
}

func (o *SecondChanceOffer) TableName() string { return SecondChanceOfferTableName }

func (o *SecondChanceOffer) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         o.Id,
		"auction_id": o.AuctionId,
		"seller_id":  o.SellerId,
		"buyer_id":   o.BuyerId,
		"bid_id":     o.BidId,
		"status":     o.Status,
		"expired_at": o.ExpiredAt,
		"created_at": o.CreatedAt,
		"updated_at": o.UpdatedAt,
	}
}

type SecondChanceOfferQueryOption struct {
	QueryOption

	AuctionId *int64
	SellerId  *int64
	BuyerId   *int64
	Status    *string
}

var _ PrepareOption = &SecondChanceOfferQueryOption{}

func (o *SecondChanceOfferQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *SecondChanceOfferQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"sco." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
