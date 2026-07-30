package model

import "auction-service/data_type"

const UserStrikeTableName = "user_strikes"

type UserStrike struct {
	Id           int64                  `db:"id"`
	BidderId     int64                  `db:"bidder_id"`
	AuctionId    int64                  `db:"auction_id"`
	SellerId     int64                  `db:"seller_id"`
	StrikeReason string                 `db:"strike_reason"`
	Status       string                 `db:"status"`
	ExpiredAt    data_type.NullDateTime `db:"expired_at"`
	Timestamp
}

func (s *UserStrike) TableName() string { return UserStrikeTableName }

func (s *UserStrike) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            s.Id,
		"bidder_id":     s.BidderId,
		"auction_id":    s.AuctionId,
		"seller_id":     s.SellerId,
		"strike_reason": s.StrikeReason,
		"status":        s.Status,
		"expired_at":    s.ExpiredAt,
		"created_at":    s.CreatedAt,
		"updated_at":    s.UpdatedAt,
	}
}

type UserStrikeQueryOption struct {
	QueryOption

	BidderId *int64
	Status   *string
}

var _ PrepareOption = &UserStrikeQueryOption{}

func (o *UserStrikeQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *UserStrikeQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"us." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
