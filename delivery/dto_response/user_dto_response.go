package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
)

// UserResponse is a single user in list endpoints.
type UserResponse struct {
	Id                      int64              `json:"id"                          example:"1"`
	Fullname                string             `json:"fullname"                    example:"John Doe"`
	Email                   string             `json:"email"                       example:"user@example.com"`
	Nik                     *string            `json:"nik,omitempty"`
	Birth                   data_type.DateTime `json:"birth"                       example:"1990-01-15T00:00:00Z"`
	Balance                 float64            `json:"balance"                     example:"100000.00"`
	Gender                  *string            `json:"gender,omitempty"            example:"MALE"`
	BankAccountNumber       *string            `json:"bank_account_number,omitempty"`
	IdentityImageLink       *string            `json:"identity_image_link,omitempty"`
	SelfieIdentityImageLink *string            `json:"selfie_identity_image_link,omitempty"`
	IsVerified              bool               `json:"is_verified"                 example:"false"`
	IsDeleted               bool               `json:"is_deleted"                  example:"false"`
	Roles                   []UserRoleResponse `json:"roles"`
	Timestamp
} // @name UserResponse

// UserRoleResponse is a single user role in responses.
type UserRoleResponse struct {
	Id     int64  `json:"id"      example:"1"`
	Role   string `json:"role"    example:"ADMIN"`
	UserId int64  `json:"user_id" example:"2"`
	Timestamp
} // @name UserRoleResponse

func NewUserResponse(_ context.Context, u model.User) UserResponse {
	r := UserResponse{
		Id:                      u.Id,
		Fullname:                u.Fullname,
		Email:                   u.Email,
		Nik:                     u.Nik,
		Birth:                   u.Birth,
		Gender:                  u.Gender,
		Balance:                 u.Balance,
		BankAccountNumber:       u.BankAccountNumber,
		IdentityImageLink:       u.IdentityImageLink,
		SelfieIdentityImageLink: u.SelfieIdentityImageLink,
		IsVerified:              u.IsVerified,
		IsDeleted:               u.IsDeleted,
		Timestamp:               Timestamp(u.Timestamp),
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
