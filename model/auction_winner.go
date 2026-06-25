package model

const AuctionWinnerTableName = "auction_winners"

type AuctionWinner struct {
	Id           int64  `db:"id"`
	AuctionId    int64  `db:"auction_id"`
	AuctionBidId *int64 `db:"auction_bid_id"`
	Status       string `db:"status"`
	Timestamp

	// relations
	Auction    *Auction    `db:"-"`
	AuctionBid *AuctionBid `db:"-"`
}

func (w *AuctionWinner) TableName() string { return AuctionWinnerTableName }

func (w *AuctionWinner) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             w.Id,
		"auction_id":     w.AuctionId,
		"auction_bid_id": w.AuctionBidId,
		"status":         w.Status,
		"created_at":     w.CreatedAt,
		"updated_at":     w.UpdatedAt,
	}
}

type AuctionWinnerQueryOption struct {
	QueryOption

	AuctionId *int64
	Status    *string
}

var _ PrepareOption = &AuctionWinnerQueryOption{}

func (o *AuctionWinnerQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *AuctionWinnerQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"aw." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
