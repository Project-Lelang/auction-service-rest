package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
)

// UserResponse is a single user in list endpoints.
type UserResponse struct {
	Id                string             `json:"id"                          example:"550e8400-e29b-41d4-a716-446655440000"`
	Fullname          string             `json:"fullname"                    example:"John Doe"`
	Phone             string             `json:"phone"                       example:"+6281234567890"`
	Nik               *string            `json:"nik,omitempty"`
	Birth             data_type.DateTime `json:"birth"                       example:"1990-01-15T00:00:00Z"`
	Gender            *string            `json:"gender,omitempty"            example:"MALE"`
	BankAccountNumber *string            `json:"bank_account_number,omitempty"`
	IsVerified        bool               `json:"is_verified"                 example:"false"`
	IsDeleted         bool               `json:"is_deleted"                  example:"false"`
	Roles             []UserRoleResponse `json:"roles"`
	Timestamp
} // @name UserResponse

// UserRoleResponse is a single user role in responses.
type UserRoleResponse struct {
	Id     string `json:"id"      example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string `json:"role"    example:"ADMIN"`
	UserId string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Timestamp
} // @name UserRoleResponse

func NewUserResponse(_ context.Context, u model.User) UserResponse {
	r := UserResponse{
		Id:                u.Id,
		Fullname:          u.Fullname,
		Phone:             u.Phone,
		Nik:               u.Nik,
		Birth:             u.Birth,
		Gender:            u.Gender,
		BankAccountNumber: u.BankAccountNumber,
		IsVerified:        u.IsVerified,
		IsDeleted:         u.IsDeleted,
		Timestamp:         Timestamp(u.Timestamp),
	}

	for _, role := range u.Roles {
		r.Roles = append(r.Roles, UserRoleResponse{
			Id:        role.Id,
			Role:      role.Role,
			UserId:    role.UserId,
			Timestamp: Timestamp(role.Timestamp),
		})
	}

	return r
}
