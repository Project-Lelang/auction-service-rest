package model

const WithdrawalRequestTableName = "withdrawal_requests"

const (
	WithdrawalRequestStatusRequested = "REQUESTED"
	WithdrawalRequestStatusCompleted = "COMPLETED"
)

type WithdrawalRequest struct {
	Id              int64   `db:"id" json:"id"`
	UserId          string  `db:"user_id" json:"user_id"`
	ValidatorUserId *string `db:"validator_user_id" json:"validator_user_id"`

	Amount float64 `db:"amount" json:"amount"`

	Status string `db:"status" json:"status"`

	// relation
	User *WithdrawalRequestUser `db:"-" json:"user,omitempty"`

	Timestamp
}

type WithdrawalRequestUser struct {
	Id string `json:"id"`

	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`

	BankName          *string `json:"bank_name"`
	BankAccountName   *string `json:"bank_account_name"`
	BankAccountNumber *string `json:"bank_account_number"`
}

func (w *WithdrawalRequest) TableName() string {
	return WithdrawalRequestTableName
}

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

type WithdrawalRequestQueryOption struct {
	QueryOption

	UserId          *string
	ValidatorUserId *string
	Status          *string
}
