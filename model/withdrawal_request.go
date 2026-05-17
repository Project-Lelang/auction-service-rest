package model

const WithdrawalRequestTableName = "withdrawal_requests"

type WithdrawalRequest struct {
	Id              string  `db:"id"`
	UserId          string  `db:"user_id"`
	ValidatorUserId *string `db:"validator_user_id"`
	Amount          float64 `db:"amount"`
	Status          string  `db:"status"`
	Timestamp

	// relations
	User          *User `db:"-"`
	ValidatorUser *User `db:"-"`
}

func (w *WithdrawalRequest) TableName() string { return WithdrawalRequestTableName }

func (w *WithdrawalRequest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                w.Id,
		"user_id":           w.UserId,
		"validator_user_id": w.ValidatorUserId,
		"amount":            w.Amount,
		"status":            w.Status,
		"created_at":        w.CreatedAt,
		"updated_at":        w.UpdatedAt,
	}
}
