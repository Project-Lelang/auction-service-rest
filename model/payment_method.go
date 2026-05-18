package model

const PaymentMethodTableName = "payment_methods"

type PaymentMethod struct {
	Id       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Code     string `db:"code" json:"code"`
	Type     string `db:"type" json:"type"`
	IsActive bool   `db:"is_active" json:"is_active"`
	Timestamp
}

func (p *PaymentMethod) TableName() string {
	return PaymentMethodTableName
}

func (p *PaymentMethod) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         p.Id,
		"name":       p.Name,
		"code":       p.Code,
		"type":       p.Type,
		"is_active":  p.IsActive,
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
}

type PaymentMethodQueryOption struct {
	QueryOption

	Name     *string `json:"name"`
	Code     *string `json:"code"`
	Type     *string `json:"type"`
	IsActive *bool   `json:"is_active"`
}
