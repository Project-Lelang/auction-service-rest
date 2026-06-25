package dto_request

// ProductFetchSorts defines sortable fields for the product fetch endpoint.
// The anonymous-struct layout must stay identical to model.Sorts for direct
// type-conversion: model.Sorts(request.Sorts).
type ProductFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id name status condition created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                  example:"desc"`
} // @name ProductFetchSorts

type ProductFetchRequest struct {
	PaginationRequest
	Sorts     ProductFetchSorts `json:"sorts"     validate:"dive"`
	Status    *string           `json:"status"    validate:"omitempty,oneof=DRAFT REQUEST VERIFIED REJECTED ON_BIDS COMPLETED" example:"VERIFIED"`
	Condition *string           `json:"condition" validate:"omitempty,oneof=NEW PRELOVED" example:"NEW"`
	Search    *string           `json:"search"    validate:"omitempty,max=255"                                            example:"laptop"`
} // @name ProductFetchRequest

type ProductGetRequest struct {
	ProductId int64 `json:"-" swaggerignore:"true"`
} // @name ProductGetRequest
