package dto_response

import (
	"context"

	"auction-service/model"
)

// UserAddressResponse represents a single user address in API responses.
type UserAddressResponse struct {
	Id             int64    `json:"id"                    example:"1"`
	UserId         int64    `json:"user_id"               example:"2"`
	Label          string   `json:"label"                 example:"Rumah"`
	RecipientName  string   `json:"recipient_name"        example:"John Doe"`
	Phone          string   `json:"phone"                 example:"08123456789"`
	CityId         string   `json:"city_id"               example:"151"`
	CityName       string   `json:"city_name"             example:"Bandung"`
	ProvinceName   string   `json:"province_name"         example:"Jawa Barat"`
	Address        string   `json:"address"               example:"Jl. Sudirman No. 1"`
	PostalCode     string   `json:"postal_code"           example:"40111"`
	BiteshipAreaId string   `json:"biteship_area_id"       example:"IDNP6IDNC148IDND843IDZ12250"`
	Latitude       *float64 `json:"latitude,omitempty"     example:"-6.2000000"`
	Longitude      *float64 `json:"longitude,omitempty"    example:"106.8166667"`
	IsDefault      bool     `json:"is_default"`
	Timestamp
} // @name UserAddressResponse

func NewUserAddressResponse(_ context.Context, a model.UserAddress) UserAddressResponse {
	return UserAddressResponse{
		Id:             a.Id,
		UserId:         a.UserId,
		Label:          a.Label,
		RecipientName:  a.RecipientName,
		Phone:          a.Phone,
		CityId:         a.CityId,
		CityName:       a.CityName,
		ProvinceName:   a.ProvinceName,
		Address:        a.Address,
		PostalCode:     a.PostalCode,
		BiteshipAreaId: a.BiteshipAreaId,
		Latitude:       a.Latitude,
		Longitude:      a.Longitude,
		IsDefault:      a.IsDefault,
		Timestamp:      Timestamp(a.Timestamp),
	}
}
