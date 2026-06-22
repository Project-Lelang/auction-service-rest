package model

const ProductStatusHistoryTableName = "product_status_histories"

type ProductStatusHistory struct {
	Id        int64   `db:"id"`
	ProductId int64   `db:"product_id"`
	Status    string  `db:"status"`
	Message   *string `db:"message"`
	Timestamp
}

func (h *ProductStatusHistory) TableName() string { return ProductStatusHistoryTableName }

func (h *ProductStatusHistory) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         h.Id,
		"product_id": h.ProductId,
		"status":     h.Status,
		"message":    h.Message,
		"created_at": h.CreatedAt,
		"updated_at": h.UpdatedAt,
	}
}
