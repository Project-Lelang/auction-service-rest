package dto_request

type WithdrawalRequestCreateRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0" example:"150000"`
}

type WithdrawalRequestFetchRequest struct {
	PaginationRequest

	Status *string `form:"status" validate:"omitempty,oneof=REQUESTED COMPLETED"`
}

type WithdrawalRequestCompleteRequest struct {
}
