package dto_request

type PaginationRequest struct {
	Page  *int `json:"page" validate:"omitempty,gte=1"`
	Limit *int `json:"limit" validate:"omitempty,gte=1,lte=100"`
}
