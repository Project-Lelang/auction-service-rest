package dto_response

import "auction-service/model"

// RoleRequestResponse represents a single role request in API responses.
type RoleRequestResponse struct {
	Id      int64   `json:"id"              example:"1"`
	UserId  int64   `json:"user_id"         example:"2"`
	Status  string  `json:"status"          example:"REQUESTED"`
	Role    string  `json:"role"            example:"BIDDER"`
	Message *string `json:"message,omitempty" example:"Please review my request"`
	Timestamp
} // @name RoleRequestResponse

func NewRoleRequestResponse(roleRequest model.RoleRequest) RoleRequestResponse {
	return RoleRequestResponse{
		Id:      roleRequest.Id,
		UserId:  roleRequest.UserId,
		Status:  roleRequest.Status,
		Role:    roleRequest.Role,
		Message: roleRequest.Message,
		Timestamp: Timestamp{
			CreatedAt: roleRequest.CreatedAt,
			UpdatedAt: roleRequest.UpdatedAt,
		},
	}
}
