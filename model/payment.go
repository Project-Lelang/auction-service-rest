package model

import "auction-service/data_type"

const PaymentTableName = "payments"

type Payment struct {
	Id              string                 `db:"id"`
	AuctionId       string                 `db:"auction_id"`
	UserId          string                 `db:"user_id"`
	PaymentMethodId *string                `db:"payment_method_id"`
	Amount          float64                `db:"amount"`
	Status          string                 `db:"status"`
	SnapUrl         *string                `db:"snap_url"`
	SnapToken       *string                `db:"snap_token"`
	ExpiredAt       data_type.NullDateTime `db:"expired_at"`
	Timestamp

	// relations
	Auction       *Auction       `db:"-"`
	User          *User          `db:"-"`
	PaymentMethod *PaymentMethod `db:"-"`
}

func (p *Payment) TableName() string { return PaymentTableName }

func (p *Payment) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                p.Id,
		"auction_id":        p.AuctionId,
		"user_id":           p.UserId,
		"payment_method_id": p.PaymentMethodId,
		"amount":            p.Amount,
		"status":            p.Status,
		"snap_url":          p.SnapUrl,
		"snap_token":        p.SnapToken,
		"expired_at":        p.ExpiredAt,
		"created_at":        p.CreatedAt,
		"updated_at":        p.UpdatedAt,
	}
}

type PaymentQueryOption struct {
	QueryOption

	AuctionId *string
	UserId    *string
	Status    *string
}

var _ PrepareOption = &PaymentQueryOption{}

func (o *PaymentQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *PaymentQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"p." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
