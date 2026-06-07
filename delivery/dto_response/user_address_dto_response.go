package dto_response

import (
	"context"

	"auction-service/model"
)

// UserAddressResponse represents a single user address in API responses.
type UserAddressResponse struct {
	Id             string `json:"id"                    example:"550e8400-e29b-41d4-a716-446655440000"`
	UserId         string `json:"user_id"               example:"550e8400-e29b-41d4-a716-446655440001"`
	Label          string `json:"label"                 example:"Rumah"`
	RecipientName  string `json:"recipient_name"        example:"John Doe"`
	Phone          string `json:"phone"                 example:"08123456789"`
	CityId         string `json:"city_id"               example:"151"`
	CityName       string `json:"city_name"             example:"Bandung"`
	ProvinceName   string `json:"province_name"         example:"Jawa Barat"`
	Address        string `json:"address"               example:"Jl. Sudirman No. 1"`
	PostalCode     string `json:"postal_code"           example:"40111"`
	BiteshipAreaId string `json:"biteship_area_id"       example:"IDNP6IDNC148IDND843IDZ12250"`
	IsDefault      bool   `json:"is_default"`
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
		IsDefault:      a.IsDefault,
		Timestamp:      Timestamp(a.Timestamp),
	}
}
