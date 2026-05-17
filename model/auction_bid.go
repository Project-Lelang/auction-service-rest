package model

const AuctionBidTableName = "auction_bids"

type AuctionBid struct {
	Id        string  `db:"id"`
	UserId    string  `db:"user_id"`
	AuctionId string  `db:"auction_id"`
	Amount    float64 `db:"amount"`
	Timestamp

	// relations
	User    *User    `db:"-"`
	Auction *Auction `db:"-"`
}

func (b *AuctionBid) TableName() string { return AuctionBidTableName }

func (b *AuctionBid) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         b.Id,
		"user_id":    b.UserId,
		"auction_id": b.AuctionId,
		"amount":     b.Amount,
		"created_at": b.CreatedAt,
		"updated_at": b.UpdatedAt,
	}
}

type AuctionBidQueryOption struct {
	QueryOption

	UserId    *string
	AuctionId *string
}

var _ PrepareOption = &AuctionBidQueryOption{}

func (o *AuctionBidQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *AuctionBidQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"ab." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
