package constant

const (
	// auth
	LanguageAuthInvalidCredentials   = "AUTH.INVALID_CREDENTIALS"
	LanguageAuthWrongPassword        = "AUTH.WRONG_PASSWORD"
	LanguageAuthAccountNotRegistered = "AUTH.ACCOUNT_NOT_REGISTERED"
	LanguageAuthTokenExpired         = "AUTH.TOKEN_EXPIRED"
	LanguageAuthTokenInvalid         = "AUTH.TOKEN_INVALID"
	LanguageAuthRefreshTokenExpired  = "AUTH.REFRESH_TOKEN_EXPIRED"
	LanguageAuthRefreshTokenInvalid  = "AUTH.REFRESH_TOKEN_INVALID"
	LanguageAuthLoggedOut            = "AUTH.LOGGED_OUT"

	// user
	LanguageUserNotFound          = "USER.NOT_FOUND"
	LanguageUserEmailAlreadyExist = "USER.EMAIL_ALREADY_EXIST"
	LanguageUserInvalidPassword   = "USER.INVALID_PASSWORD"

	// system
	LanguageSystemUnauthorized          = "SYSTEM.UNAUTHORIZED"
	LanguageSystemForbidden             = "SYSTEM.FORBIDDEN"
	LanguageSystemInvalidRequestPayload = "SYSTEM.INVALID_REQUEST_PAYLOAD"
	LanguageSystemInternalServerError   = "SYSTEM.INTERNAL_SERVER_ERROR"
	LanguageSystemNotFound              = "SYSTEM.NOT_FOUND"
	LanguageSystemMustBeAValidUuid      = "SYSTEM.MUST_BE_A_VALID_UUID"
)
