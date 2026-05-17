package model

import "auction-service/data_type"

const UserTableName = "users"

type User struct {
	Id                      string             `db:"id"`
	Fullname                string             `db:"fullname"`
	Phone                   string             `db:"phone"`
	Nik                     *string            `db:"nik"`
	Birth                   data_type.DateTime `db:"birth"`
	Gender                  *string            `db:"gender"`
	BankAccountNumber       *string            `db:"bank_account_number"`
	IdentityImagePath       *string            `db:"identity_image_path"`
	SelfieIdentityImagePath *string            `db:"selfie_identity_image_path"`
	Balance                 float64            `db:"balance"`
	IsVerified              bool               `db:"is_verified"`
	IsDeleted               bool               `db:"is_deleted"`
	Password                string             `db:"password"`
	Timestamp

	// relations
	Roles []UserRole `db:"-"`

	// computed
	IdentityImageLink       *string `db:"-"`
	SelfieIdentityImageLink *string `db:"-"`
}

func (u *User) TableName() string { return UserTableName }

func (u *User) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":                         u.Id,
		"fullname":                   u.Fullname,
		"phone":                      u.Phone,
		"nik":                        u.Nik,
		"birth":                      u.Birth,
		"gender":                     u.Gender,
		"bank_account_number":        u.BankAccountNumber,
		"identity_image_path":        u.IdentityImagePath,
		"selfie_identity_image_path": u.SelfieIdentityImagePath,
		"balance":                    u.Balance,
		"is_verified":                u.IsVerified,
		"is_deleted":                 u.IsDeleted,
		"password":                   u.Password,
		"created_at":                 u.CreatedAt,
		"updated_at":                 u.UpdatedAt,
	}
}

type UserQueryOption struct {
	QueryOption

	Role *string
}

var _ PrepareOption = &UserQueryOption{}

// SetDefaultSorts overrides the base default to use id ASC.
func (o *UserQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "id", Direction: "asc"}}
	}
}

// TranslateSorts prefixes every sort field with the users table alias "u.".
func (o *UserQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"u." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
