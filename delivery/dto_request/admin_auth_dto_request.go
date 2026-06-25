package dto_request

type AdminAuthLoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"admin@example.com"`
	Password string `json:"password" validate:"required,min=8,max=255" example:"password123"`
} // @name AuthLoginRequest
