package model

import "time"

const UserTableName = "users"

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "superadmin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleUser       UserRole = "user"
)

type User struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Role      UserRole  `db:"role"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u *User) TableName() string {
	return UserTableName
}

func (u *User) TableIds() []string {
	return []string{"id"}
}

func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.Id,
		"name":       u.Name,
		"email":      u.Email,
		"password":   u.Password,
		"role":       u.Role,
		"is_active":  u.IsActive,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}

type UserQueryOption struct {
	IsCount bool
	IdIn    []string
	Phrase  *string
	Page    *int
	Limit   *int
}

type UserAccessToken struct {
	Id        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Token     string    `db:"token"`
	ExpiredAt time.Time `db:"expired_at"`
	CreatedAt time.Time `db:"created_at"`
}

const UserAccessTokenTableName = "user_access_tokens"

func (u *UserAccessToken) TableName() string {
	return UserAccessTokenTableName
}

func (u *UserAccessToken) TableIds() []string {
	return []string{"id"}
}

func (u *UserAccessToken) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.Id,
		"user_id":    u.UserId,
		"token":      u.Token,
		"expired_at": u.ExpiredAt,
		"created_at": u.CreatedAt,
	}
}
