package model

const UserAddressTableName = "user_addresses"

type UserAddress struct {
	Id             int64    `db:"id"`
	UserId         int64    `db:"user_id"`
	Label          string   `db:"label"`
	RecipientName  string   `db:"recipient_name"`
	Phone          string   `db:"phone"`
	CityId         string   `db:"city_id"`
	CityName       string   `db:"city_name"`
	ProvinceName   string   `db:"province_name"`
	Address        string   `db:"address"`
	PostalCode     string   `db:"postal_code"`
	BiteshipAreaId string   `db:"biteship_area_id"`
	Latitude       *float64 `db:"latitude"`
	Longitude      *float64 `db:"longitude"`
	IsDefault      bool     `db:"is_default"`
	Timestamp

	// relations
	User *User `db:"-"`
}

func (a *UserAddress) TableName() string { return UserAddressTableName }

func (a *UserAddress) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":               a.Id,
		"user_id":          a.UserId,
		"label":            a.Label,
		"recipient_name":   a.RecipientName,
		"phone":            a.Phone,
		"city_id":          a.CityId,
		"city_name":        a.CityName,
		"province_name":    a.ProvinceName,
		"address":          a.Address,
		"postal_code":      a.PostalCode,
		"biteship_area_id": a.BiteshipAreaId,
		"latitude":         a.Latitude,
		"longitude":        a.Longitude,
		"is_default":       a.IsDefault,
		"created_at":       a.CreatedAt,
		"updated_at":       a.UpdatedAt,
	}
}

type UserAddressQueryOption struct {
	QueryOption

	UserId *int64
}

var _ PrepareOption = &UserAddressQueryOption{}

func (o *UserAddressQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

func (o *UserAddressQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"ua." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
