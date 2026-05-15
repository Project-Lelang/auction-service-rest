package dto_response

// AuthTokenResponse is returned by login endpoints.
type AuthTokenResponse struct {
	Token string `json:"token" example:"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
} // @name AuthTokenResponse

func NewAuthTokenResponse(token string) AuthTokenResponse {
	return AuthTokenResponse{Token: token}
}
