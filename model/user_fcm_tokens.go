package model

const UserFcmTokenTableName = "user_fcm_tokens"

type UserFcmToken struct {
	Id       int64  `db:"id"`
	UserId   int64  `db:"user_id"`
	FcmToken string `db:"fcm_token"`
	Timestamp
}

func (u *UserFcmToken) TableName() string { return UserFcmTokenTableName }

func (u *UserFcmToken) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.Id,
		"user_id":    u.UserId,
		"fcm_token":  u.FcmToken,
		"created_at": u.CreatedAt,
		"updated_at": u.UpdatedAt,
	}
}

type UserFcmTokenQueryOption struct {
	QueryOption
}

var _ PrepareOption = &UserFcmTokenQueryOption{}

// SetDefaultSorts overrides the base default to use id ASC.
func (o *UserFcmTokenQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "id", Direction: "asc"}}
	}
}

// TranslateSorts prefixes every sort field with the user_fcm_tokens table alias "u.".
func (o *UserFcmTokenQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"u." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
