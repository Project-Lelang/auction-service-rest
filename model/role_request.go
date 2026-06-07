package model

const RoleRequestTableName = "role_requests"

type RoleRequest struct {
	Id      int64   `db:"id" json:"id"`
	UserId  string  `db:"user_id" json:"user_id"`
	Status  string  `db:"status" json:"status"`
	Role    string  `db:"role" json:"role"`
	Message *string `db:"message" json:"message"`

	Timestamp

	User *RoleRequestUser `db:"-" json:"user,omitempty"`
}

type RoleRequestUser struct {
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`

	Nik *string `json:"nik"`

	BankAccountNumber *string `json:"bank_account_number"`
	BankAccountName   *string `json:"bank_account_name"`
	BankName          *string `json:"bank_name"`

	IdentityImagePath       *string `json:"identity_image_path"`
	SelfieIdentityImagePath *string `json:"selfie_identity_image_path"`
}

func (r *RoleRequest) TableName() string {
	return RoleRequestTableName
}

func (r *RoleRequest) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         r.Id,
		"user_id":    r.UserId,
		"status":     r.Status,
		"role":       r.Role,
		"message":    r.Message,
		"created_at": r.CreatedAt,
		"updated_at": r.UpdatedAt,
	}
}

type RoleRequestQueryOption struct {
	QueryOption

	UserId *string
	Status *string
	Role   *string
}
