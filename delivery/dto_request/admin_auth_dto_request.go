package dto_request

type AdminAuthLoginRequest struct {
	Phone    string `json:"phone" validate:"required,max=20" example:"+6281234567890"`
	Password string `json:"password" validate:"required,min=8,max=255" example:"password123"`
} // @name AuthLoginRequest
