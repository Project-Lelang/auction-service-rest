package dto_request

type PaginationRequest struct {
	Page  *int `form:"page" json:"page" validate:"omitempty,gte=1"`
	Limit *int `form:"limit" json:"limit" validate:"omitempty,gte=1,lte=100"`
}
