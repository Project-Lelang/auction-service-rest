package dto_request

type PaymentMethodFetchRequest struct {
	PaginationRequest

	Name     *string `form:"name" validate:"omitempty,max=100"`
	Code     *string `form:"code" validate:"omitempty,max=50"`
	Type     *string `form:"type" validate:"omitempty,max=50"`
	IsActive *bool   `form:"is_active" validate:"omitempty"`
}

type PaymentMethodCreateRequest struct {
	Name string `json:"name" validate:"required,max=100" example:"BCA Virtual Account"`
	Code string `json:"code" validate:"required,max=50" example:"bca_va"`
	Type string `json:"type" validate:"required,max=50" example:"BANK_TRANSFER"`
}

type PaymentMethodUpdateRequest struct {
	Name     *string `json:"name" validate:"omitempty,max=100" example:"BCA Virtual Account Baru"`
	Code     *string `json:"code" validate:"omitempty,max=50" example:"bca_va_new"`
	Type     *string `json:"type" validate:"omitempty,max=50" example:"BANK_TRANSFER"`
	IsActive *bool   `json:"is_active" validate:"omitempty" example:"true"`
}

type PaymentMethodToggleStatusRequest struct {
	IsActive bool `json:"is_active" validate:"required" example:"false"`
}
