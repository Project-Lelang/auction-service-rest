package dto_response

import "auction-service/model"

// WithdrawalRequestResponse represents a single withdrawal request in API responses.
type WithdrawalRequestResponse struct {
	Id              string  `json:"id"                          example:"550e8400-e29b-41d4-a716-446655440000"`
	UserId          string  `json:"user_id"                     example:"550e8400-e29b-41d4-a716-446655440001"`
	ValidatorUserId *string `json:"validator_user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Amount          float64 `json:"amount"                      example:"500000"`
	Status          string  `json:"status"                      example:"REQUESTED"`
	Timestamp
} // @name WithdrawalRequestResponse

func NewWithdrawalRequestResponse(wr model.WithdrawalRequest) WithdrawalRequestResponse {
	return WithdrawalRequestResponse{
		Id:              wr.Id,
		UserId:          wr.UserId,
		ValidatorUserId: wr.ValidatorUserId,
		Amount:          wr.Amount,
		Status:          wr.Status,
		Timestamp: Timestamp{
			CreatedAt: wr.CreatedAt,
			UpdatedAt: wr.UpdatedAt,
		},
	}
}
