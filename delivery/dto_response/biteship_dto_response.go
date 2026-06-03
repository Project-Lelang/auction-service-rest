package dto_response

import "auction-service/infrastructure"

// BiteshipAreaResponse wraps a single Biteship area (district-level).
type BiteshipAreaResponse struct {
	Id         string `json:"id"            example:"IDNP6IDNC148IDND843IDZ12250"`
	Name       string `json:"name"          example:"Pesanggrahan, Jakarta Selatan, DKI Jakarta. 12250"`
	Province   string `json:"province"      example:"DKI Jakarta"`
	City       string `json:"city"          example:"Jakarta Selatan"`
	District   string `json:"district"      example:"Pesanggrahan"`
	PostalCode int    `json:"postal_code"   example:12250`
} // @name BiteshipAreaResponse

func NewBiteshipAreaResponse(a infrastructure.BiteshipArea) BiteshipAreaResponse {
	return BiteshipAreaResponse{
		Id:         a.Id,
		Name:       a.Name,
		Province:   a.Province,
		City:       a.City,
		District:   a.District,
		PostalCode: a.PostalCode,
	}
}
