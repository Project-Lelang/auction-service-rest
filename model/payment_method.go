package model

const PaymentMethodTableName = "payment_methods"

type PaymentMethod struct {
	Id       string `db:"id"`
	Code     string `db:"code"`
	Type     string `db:"type"`
	Name     string `db:"name"`
	IsActive bool   `db:"is_active"`
	Timestamp
}

func (p *PaymentMethod) TableName() string { return PaymentMethodTableName }

func (p *PaymentMethod) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         p.Id,
		"code":       p.Code,
		"type":       p.Type,
		"name":       p.Name,
		"is_active":  p.IsActive,
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
}

type PaymentMethodQueryOption struct {
	QueryOption

	IsActive *bool
	Type     *string
}

var _ PrepareOption = &PaymentMethodQueryOption{}

func (o *PaymentMethodQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *PaymentMethodQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"pm." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
