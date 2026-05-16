package dto_request

type OwnProfileUpdateRequest struct {
	Fullname          string  `json:"fullname"            validate:"required,max=255"          example:"John Doe"`
	Phone             string  `json:"phone"               validate:"required,max=20"           example:"+6281234567890"`
	Nik               *string `json:"nik"                 validate:"omitempty,max=50"          example:"3201010101010001"`
	Birth             string  `json:"birth"               validate:"required"                  example:"1990-01-15"`
	Gender            *string `json:"gender"              validate:"omitempty,oneof=MALE FEMALE" example:"MALE"`
	BankAccountNumber *string `json:"bank_account_number" validate:"omitempty,max=50"          example:"1234567890"`
} // @name OwnProfileUpdateRequest

// OwnProductFetchSorts defines sortable fields for the own product fetch endpoint.
// The anonymous-struct layout must stay identical to model.Sorts for direct
// type-conversion: model.Sorts(request.Sorts).
type OwnProductFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id name status condition created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                  example:"desc"`
} // @name OwnProductFetchSorts

type OwnProductFetchRequest struct {
	PaginationRequest
	Sorts     OwnProductFetchSorts `json:"sorts"     validate:"dive"`
	Status    *string              `json:"status"    validate:"omitempty,oneof=DRAFT REQUEST VERIFIED REJECTED ON_BIDS COMPLETED" example:"VERIFIED"`
	Condition *string              `json:"condition" validate:"omitempty,oneof=NEW PRELOVED" example:"NEW"`
	Search    *string              `json:"search"    validate:"omitempty,max=255"                                            example:"laptop"`
} // @name OwnProductFetchRequest

type OwnProductFetchStatusHistoriesRequest struct {
	ProductId string `json:"-" swaggerignore:"true"`
} // @name OwnProductFetchStatusHistoriesRequest

type OwnProductGetRequest struct {
	ProductId string `json:"-" swaggerignore:"true"`
} // @name OwnProductGetRequest

type OwnProductUpdateRequest struct {
	Name        string  `json:"name"        validate:"required,max=255"            example:"Vintage Camera"`
	Description *string `json:"description" validate:"omitempty,max=2000"          example:"A beautiful vintage camera"`
	Condition   string  `json:"condition"   validate:"required,oneof=NEW PRELOVED" example:"NEW"`

	ProductId string `json:"-" swaggerignore:"true"`
} // @name OwnProductUpdateRequest

type OwnProductRequestRequest struct {
	ProductId string `json:"-" swaggerignore:"true"`
} // @name OwnProductRequestRequest
