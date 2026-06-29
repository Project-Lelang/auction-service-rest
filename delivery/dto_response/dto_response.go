package dto_response

import (
	"net/http"

	"auction-service/constant"
	"auction-service/data_type"
	"auction-service/util"
)

type Response struct {
	Data interface{} `json:"data"`
} // @name Response

type SuccessResponse struct {
	Message string `json:"message" example:"OK"`
} // @name SuccessResponse

// DataResponse wraps a single entity in a named map so the JSON key documents
// the entity type (e.g. {"product": {...}}).
type DataResponse map[string]interface{} // @name DataResponse

// Timestamp is embedded in every response struct that mirrors a model with timestamps.
type Timestamp struct {
	CreatedAt data_type.DateTime `json:"created_at"`
	UpdatedAt data_type.DateTime `json:"updated_at"`
} // @name Timestamp

// PaginationResponse is returned by all paginated list endpoints.
type PaginationResponse struct {
	Total int         `json:"total" example:"24"`
	Page  *int        `json:"page"  example:"1"`
	Limit *int        `json:"limit" example:"10"`
	Nodes interface{} `json:"nodes"`
} // @name PaginationResponse

// NewPaginationResponse builds a PaginationResponse, applying default page/limit
// values when the caller passes nil pointers.
func NewPaginationResponse(nodes interface{}, total int, pageP *int, limitP *int) PaginationResponse {
	page := constant.PaginationDefaultPage
	if pageP != nil {
		page = *pageP
	}

	limit := constant.PaginationDefaultLimit
	if limitP != nil {
		limit = *limitP
	}

	return PaginationResponse{
		Nodes: nodes,
		Total: total,
		Page:  util.Pointer(page),
		Limit: util.Pointer(limit),
	}
}

type Error struct {
	Domain  string `json:"domain"`
	Message string `json:"message"`
} // @name Error

type ErrorResponse struct {
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Contents []string `json:"-" swaggerignore:"true"`
	Errors   []Error  `json:"errors"`
} // @name ErrorResponse

func (e ErrorResponse) Error() string {
	return e.Message
}

func NewUnauthorizedErrorResponse(message string, contents ...string) ErrorResponse {
	return ErrorResponse{
		Code:     http.StatusUnauthorized,
		Message:  message,
		Contents: contents,
		Errors:   []Error{},
	}
}

func NewBadRequestErrorResponse(message string, contents ...string) ErrorResponse {
	return ErrorResponse{
		Code:     http.StatusBadRequest,
		Message:  message,
		Contents: contents,
		Errors:   []Error{},
	}
}

func NewForbiddenErrorResponse(message string, contents ...string) ErrorResponse {
	return ErrorResponse{
		Code:     http.StatusForbidden,
		Message:  message,
		Contents: contents,
		Errors:   []Error{},
	}
}

func NewNotFoundErrorResponse(message string, contents ...string) ErrorResponse {
	return ErrorResponse{
		Code:     http.StatusNotFound,
		Message:  message,
		Contents: contents,
		Errors:   []Error{},
	}
}

func NewConflictErrorResponse(message string, contents ...string) ErrorResponse {
	return ErrorResponse{
		Code:     http.StatusConflict,
		Message:  message,
		Contents: contents,
		Errors:   []Error{},
	}
}

func NewInternalServerErrorResponse() ErrorResponse {
	return ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: constant.LanguageSystemInternalServerError,
		Errors:  []Error{},
	}
}

func NewInternalServerErrorResponseP() *ErrorResponse {
	r := NewInternalServerErrorResponse()
	return &r
}
