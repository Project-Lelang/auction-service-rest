package dto_response

import "auction-service/model"

// RoleRequestResponse represents a single role request in API responses.
type RoleRequestResponse struct {
	Id      int64   `json:"id"              example:"1"`
	UserId  string  `json:"user_id"         example:"550e8400-e29b-41d4-a716-446655440001"`
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
