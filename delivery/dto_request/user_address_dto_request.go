package dto_request

// --------------------------------------------------------------------- user address

type UserAddressFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=label created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                    example:"desc"`
} // @name UserAddressFetchSorts

type UserAddressFetchRequest struct {
	PaginationRequest
	Sorts UserAddressFetchSorts `json:"sorts" validate:"dive"`
} // @name UserAddressFetchRequest

type UserAddressGetRequest struct {
	UserAddressId string `json:"-" swaggerignore:"true"`
} // @name UserAddressGetRequest

type UserAddressCreateRequest struct {
	Label          string `json:"label"                  validate:"required,max=100"  example:"Rumah"`
	RecipientName  string `json:"recipient_name"         validate:"required,max=255"  example:"John Doe"`
	Phone          string `json:"phone"                  validate:"required,max=20"   example:"08123456789"`
	CityId         string `json:"city_id"                validate:"omitempty,max=100" example:"151"`
	CityName       string `json:"city_name"              validate:"required,max=100"  example:"Bandung"`
	ProvinceName   string `json:"province_name"          validate:"required,max=100"  example:"Jawa Barat"`
	Address        string `json:"address"                validate:"required,max=500"  example:"Jl. Sudirman No. 1"`
	PostalCode     string `json:"postal_code"            validate:"required,max=10"   example:"40111"`
	BiteshipAreaId string `json:"biteship_area_id"       validate:"required,min=1"    example:"IDNP6IDNC148IDND843IDZ12250"`
	IsDefault      bool   `json:"is_default"`
} // @name UserAddressCreateRequest

type UserAddressUpdateRequest struct {
	Label          string `json:"label"                  validate:"required,max=100"  example:"Rumah"`
	RecipientName  string `json:"recipient_name"         validate:"required,max=255"  example:"John Doe"`
	Phone          string `json:"phone"                  validate:"required,max=20"   example:"08123456789"`
	CityId         string `json:"city_id"                validate:"omitempty,max=100" example:"151"`
	CityName       string `json:"city_name"              validate:"required,max=100"  example:"Bandung"`
	ProvinceName   string `json:"province_name"          validate:"required,max=100"  example:"Jawa Barat"`
	Address        string `json:"address"                validate:"required,max=500"  example:"Jl. Sudirman No. 1"`
	PostalCode     string `json:"postal_code"            validate:"required,max=10"   example:"40111"`
	BiteshipAreaId string `json:"biteship_area_id"       validate:"required,min=1"    example:"IDNP6IDNC148IDND843IDZ12250"`
	IsDefault      bool   `json:"is_default"`
	UserAddressId  string `json:"-" swaggerignore:"true"`
} // @name UserAddressUpdateRequest

type UserAddressDeleteRequest struct {
	UserAddressId string `json:"-" swaggerignore:"true"`
} // @name UserAddressDeleteRequest
