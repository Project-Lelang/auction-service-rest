package model

const UserRoleTableName = "user_roles"

type UserRole struct {
	Id     string `db:"id"`
	Role   string `db:"role"`
	UserId string `db:"user_id"`
	Timestamp
}

func (ur *UserRole) TableName() string { return UserRoleTableName }

func (ur *UserRole) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         ur.Id,
		"role":       ur.Role,
		"user_id":    ur.UserId,
		"created_at": ur.CreatedAt,
		"updated_at": ur.UpdatedAt,
	}
}
