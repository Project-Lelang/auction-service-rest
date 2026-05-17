package model

const RoleRequestTableName = "role_requests"

type RoleRequest struct {
	Id      string  `db:"id"`
	UserId  string  `db:"user_id"`
	Status  string  `db:"status"`
	Role    string  `db:"role"`
	Message *string `db:"message"`
	Timestamp

	// relations
	User *User `db:"-"`
}

func (r *RoleRequest) TableName() string { return RoleRequestTableName }

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
