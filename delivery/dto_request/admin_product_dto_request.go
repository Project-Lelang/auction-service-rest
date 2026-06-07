package dto_request

// AdminProductFetchSorts defines sortable fields for the admin product fetch endpoint.
// The anonymous-struct layout must stay identical to model.Sorts for direct
// type-conversion: model.Sorts(request.Sorts).
type AdminProductFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id name status condition created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                  example:"desc"`
} // @name AdminProductFetchSorts

type AdminProductFetchRequest struct {
	PaginationRequest
	Sorts     AdminProductFetchSorts `json:"sorts"     validate:"dive"`
	Status    *string                `json:"status"    validate:"omitempty,oneof=DRAFT REQUEST VERIFIED REJECTED ON_BIDS COMPLETED" example:"VERIFIED"`
	Condition *string                `json:"condition" validate:"omitempty,oneof=NEW PRELOVED" example:"NEW"`
	Search    *string                `json:"search"    validate:"omitempty,max=255"                                            example:"laptop"`
} // @name AdminProductFetchRequest

type AdminProductFetchStatusHistoriesRequest struct {
	ProductId string `json:"-" swaggerignore:"true"`
} // @name AdminProductFetchStatusHistoriesRequest

type AdminProductApproveRequest struct {
	UserId    string `uri:"userId"`
	ProductId string `uri:"productId"`
}

type AdminProductRejectRequest struct {
	UserId    string  `uri:"userId"`
	ProductId string  `uri:"productId"`
	Message   *string `json:"message" binding:"required"` // Mandatory rejection reason
}
