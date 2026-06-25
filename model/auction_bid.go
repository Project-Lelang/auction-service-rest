package model

const AuctionBidTableName = "auction_bids"

type AuctionBid struct {
	Id        int64   `db:"id"`
	UserId    int64   `db:"user_id"`
	AuctionId int64   `db:"auction_id"`
	Amount    float64 `db:"amount"`
	Timestamp

	// relations
	User    *User          `db:"-"`
	Auction *Auction       `db:"-"`
	Winner  *AuctionWinner `db:"-"`
	Payment *Payment       `db:"-"`
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

	UserId    *int64
	AuctionId *int64
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
