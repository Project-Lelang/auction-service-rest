package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
)

// UserResponse is a single user in list endpoints.
type UserResponse struct {
	Id         int64              `json:"id"          example:"1"`
	Fullname   string             `json:"fullname"    example:"John Doe"`
	IsVerified bool               `json:"is_verified" example:"false"`
	Roles      []UserRoleResponse `json:"roles"`
} // @name UserResponse

// OwnProfileResponse contains the authenticated user's complete profile data.
// Password hashes and internal storage paths are intentionally never exposed.
type OwnProfileResponse struct {
	Id                      int64                 `json:"id"                           example:"1"`
	Fullname                string                `json:"fullname"                     example:"John Doe"`
	Email                   string                `json:"email"                        example:"user@example.com"`
	Nik                     *string               `json:"nik"                           example:"3201010101010001"`
	Birth                   data_type.DateTime    `json:"birth"                        example:"1990-01-15T00:00:00Z"`
	Gender                  *string               `json:"gender"                        example:"MALE"`
	Balance                 float64               `json:"balance"                      example:"100000.00"`
	BankName                *string               `json:"bank_name"                     example:"Bank ABC"`
	BankAccountName         *string               `json:"bank_account_name"             example:"John Doe"`
	BankAccountNumber       *string               `json:"bank_account_number"           example:"1234567890"`
	IdentityImageLink       *string               `json:"identity_image_link"`
	SelfieIdentityImageLink *string               `json:"selfie_identity_image_link"`
	IsVerified              bool                  `json:"is_verified"                  example:"true"`
	IsDeleted               bool                  `json:"is_deleted"                   example:"false"`
	Roles                   []UserRoleResponse    `json:"roles"`
	RoleRequests            []RoleRequestResponse `json:"role_requests"`
	Timestamp
} // @name OwnProfileResponse

// AdminUserResponse exposes complete user data only through admin-protected APIs.
type AdminUserResponse OwnProfileResponse // @name AdminUserResponse

// UserRoleResponse is a single user role in responses.
type UserRoleResponse struct {
	Id     int64  `json:"id"      example:"1"`
	Role   string `json:"role"    example:"ADMIN"`
	UserId int64  `json:"user_id" example:"2"`
	Timestamp
} // @name UserRoleResponse

func NewUserResponse(_ context.Context, u model.User) UserResponse {
	r := UserResponse{
		Id:         u.Id,
		Fullname:   u.Fullname,
		IsVerified: u.IsVerified,
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

func NewOwnProfileResponse(_ context.Context, u model.User) OwnProfileResponse {
	r := OwnProfileResponse{
		Id:                      u.Id,
		Fullname:                u.Fullname,
		Email:                   u.Email,
		Nik:                     u.Nik,
		Birth:                   u.Birth,
		Gender:                  u.Gender,
		Balance:                 u.Balance,
		BankName:                u.BankName,
		BankAccountName:         u.BankAccountName,
		BankAccountNumber:       u.BankAccountNumber,
		IdentityImageLink:       u.IdentityImageLink,
		SelfieIdentityImageLink: u.SelfieIdentityImageLink,
		IsVerified:              u.IsVerified,
		IsDeleted:               u.IsDeleted,
		RoleRequests:            []RoleRequestResponse{},
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

	for _, roleRequest := range u.RoleRequests {
		r.RoleRequests = append(r.RoleRequests, NewRoleRequestResponse(roleRequest))
	}

	return r
}

func NewAdminUserResponse(ctx context.Context, u model.User) AdminUserResponse {
	return AdminUserResponse(NewOwnProfileResponse(ctx, u))
}
