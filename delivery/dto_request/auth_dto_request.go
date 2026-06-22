package dto_request

type AuthLoginRequest struct {
	Phone    string `json:"phone" validate:"required,max=20" example:"+6281234567890"`
	Password string `json:"password" validate:"required,min=8,max=255" example:"password123"`
} // @name AuthLoginRequest

type AuthOtpRequest struct {
	Phone string `json:"phone" validate:"required,max=20" example:"+6281234567890"`
} // @name AuthOtpRequest

type AuthRegisterRequest struct {
	Fullname string  `json:"fullname" validate:"required,max=255" example:"John Doe"`
	Phone    string  `json:"phone" validate:"required,max=20" example:"+6281234567890"`
	Birth    string  `json:"birth" validate:"required" example:"1990-01-15"`
	Gender   *string `json:"gender" validate:"omitempty,oneof=MALE FEMALE" example:"MALE"`
	Password string  `json:"password" validate:"required,min=8,max=255" example:"password123"`
	Otp      string  `json:"otp" validate:"required,len=6" example:"123456"`
} // @name AuthRegisterRequest

type AuthFcmTokenRequest struct {
	FcmToken string `json:"fcm_token" validate:"required" example:"fcm_token_123"`

	UserId int64 `json:"-"           swaggerignore:"true"`
} // @name AuthFcmTokenRequest
