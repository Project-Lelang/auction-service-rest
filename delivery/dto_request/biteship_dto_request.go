package dto_request

// --------------------------------------------------------------------- biteship

type BiteshipSearchAreasRequest struct {
	Keyword string `json:"keyword" validate:"required,min=2,max=100"`
} // @name BiteshipSearchAreasRequest
