package model

import "auction-service/data_type"

const AuctionTableName = "auctions"

type Auction struct {
	Id            int64              `db:"id"`
	Code          string             `db:"code"`
	ProductId     int64              `db:"product_id"`
	StartingPrice float64            `db:"starting_price"`
	StartTime     data_type.DateTime `db:"start_time"`
	EndTime       data_type.DateTime `db:"end_time"`
	Status        string             `db:"status"`
	Fee           float64            `db:"fee"`
	Timestamp

	// relations
	Product *Product       `db:"-"`
	Winner  *AuctionWinner `db:"-"`
	Payment *Payment       `db:"-"`
	Bids    []*AuctionBid  `db:"-"`
}

func (a *Auction) TableName() string { return AuctionTableName }

func (a *Auction) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             a.Id,
		"code":           a.Code,
		"product_id":     a.ProductId,
		"starting_price": a.StartingPrice,
		"start_time":     a.StartTime,
		"end_time":       a.EndTime,
		"status":         a.Status,
		"fee":            a.Fee,
		"created_at":     a.CreatedAt,
		"updated_at":     a.UpdatedAt,
	}
}

type AuctionQueryOption struct {
	QueryOption

	UserId   *int64
	Status   *string
	Statuses []string
}

var _ PrepareOption = &AuctionQueryOption{}

func (o *AuctionQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *AuctionQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"a." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
