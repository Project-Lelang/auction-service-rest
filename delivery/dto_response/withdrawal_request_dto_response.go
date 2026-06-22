package dto_response

import "auction-service/model"

// WithdrawalRequestResponse represents a single withdrawal request in API responses.
type WithdrawalRequestResponse struct {
	Id              int64   `json:"id"                          example:"1"`
	UserId          int64   `json:"user_id"                     example:"2"`
	ValidatorUserId *int64  `json:"validator_user_id,omitempty" example:"3"`
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
