package dto_request

type RoleRequestFetchRequest struct {
	PaginationRequest
	Status *string `form:"status" validate:"omitempty,oneof=REQUESTED APPROVED REJECTED"`
	Role   *string `form:"role" validate:"omitempty,oneof=BIDDER SELLER"`
}

type RoleRequestCreateRequest struct {
	Role string `json:"role" validate:"required,oneof=BIDDER SELLER" example:"SELLER"`
}

type RoleRequestRejectRequest struct {
	Message string `json:"message" validate:"required,max=255" example:"Dokumen pendukung kurang jelas atau blur"`
}
