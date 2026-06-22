package model

import "auction-service/data_type"

const OtpTableName = "otps"

type Otp struct {
	Id        int64              `db:"id"`
	Phone     string             `db:"phone"`
	Otp       string             `db:"otp"`
	ExpiresAt data_type.DateTime `db:"expires_at"`
	Verified  bool               `db:"verified"`
	Timestamp
}

func (o *Otp) TableName() string { return OtpTableName }

func (o *Otp) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":         o.Id,
		"phone":      o.Phone,
		"otp":        o.Otp,
		"expires_at": o.ExpiresAt,
		"verified":   o.Verified,
		"created_at": o.CreatedAt,
		"updated_at": o.UpdatedAt,
	}
}
