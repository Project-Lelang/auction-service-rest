package dto_request

// AdminUserFetchSorts defines sortable fields for the admin user fetch endpoint.
// The anonymous-struct layout must stay identical to model.Sorts for direct
// type-conversion: model.Sorts(request.Sorts).
type AdminUserFetchSorts []struct {
	Field     string `json:"field"     validate:"required,oneof=id fullname email birth gender created_at updated_at" example:"created_at"`
	Direction string `json:"direction" validate:"required,oneof=asc desc"                                      example:"desc"`
} // @name AdminUserFetchSorts

type AdminUserFetchRequest struct {
	PaginationRequest
	Sorts AdminUserFetchSorts `json:"sorts" validate:"dive"`
	Role  *string             `json:"role"  validate:"omitempty,oneof=ADMIN BIDDER SELLER SUPERADMIN" example:"BIDDER"`
} // @name AdminUserFetchRequest

type AdminUserRegisterRequest struct {
	Fullname string  `json:"fullname" validate:"required,max=255" example:"John Doe"`
	Email    string  `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Birth    string  `json:"birth" validate:"required" example:"1990-01-15"`
	Gender   *string `json:"gender" validate:"omitempty,oneof=MALE FEMALE" example:"MALE"`
	Password string  `json:"password" validate:"required,min=8,max=255" example:"password123"`
	Otp      string  `json:"otp" validate:"required,len=6" example:"123456"`
} // @name AdminUserRegisterRequest

type AdminUserCreateRequest struct {
	Fullname string  `json:"fullname" validate:"required,max=255" example:"Admin User"`
	Email    string  `json:"email" validate:"required,email,max=255" example:"admin@example.com"`
	Birth    string  `json:"birth" validate:"required" example:"1990-01-15"`
	Gender   *string `json:"gender" validate:"omitempty,oneof=MALE FEMALE" example:"MALE"`
	Password string  `json:"password" validate:"required,min=8,max=255" example:"password123"`
} // @name AdminCreateRequest

type AdminUserRoleCreateRequest struct {
	Role string `json:"role" validate:"required,oneof=SUPERADMIN ADMIN BIDDER SELLER" example:"ADMIN"`

	UserId int64 `json:"-" swaggerignore:"true"`
} // @name AdminUserRoleRequest

type AdminUserGetRequest struct {
	UserId int64 `json:"-" swaggerignore:"true"`
} // @name AdminUserGetRequest

type AdminUserDeleteRequest struct {
	UserId int64 `json:"-" swaggerignore:"true"`
} // @name AdminUserDeleteRequest

type AdminUserRoleDeleteRequest struct {
	UserId int64  `json:"-" swaggerignore:"true"`
	Role   string `json:"-" swaggerignore:"true"`
} // @name AdminUserRoleDeleteRequest
