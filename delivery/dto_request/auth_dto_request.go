package dto_request

type AuthLoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"admin@example.com"`
	Password string `json:"password" validate:"required,min=8,max=255" example:"password123"`
} // @name AuthLoginRequest

type AuthRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
} // @name AuthRefreshTokenRequest
