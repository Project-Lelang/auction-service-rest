package constant

const (
	// auth
	LanguageAuthInvalidCredentials     = "AUTH.INVALID_CREDENTIALS"
	LanguageAuthWrongPassword          = "AUTH.WRONG_PASSWORD"
	LanguageAuthAccountNotRegistered   = "AUTH.ACCOUNT_NOT_REGISTERED"
	LanguageAuthTokenExpired           = "AUTH.TOKEN_EXPIRED"
	LanguageAuthTokenInvalid           = "AUTH.TOKEN_INVALID"
	LanguageAuthRefreshTokenExpired    = "AUTH.REFRESH_TOKEN_EXPIRED"
	LanguageAuthRefreshTokenInvalid    = "AUTH.REFRESH_TOKEN_INVALID"
	LanguageAuthLoggedOut              = "AUTH.LOGGED_OUT"
	LanguageAuthOtpInvalid             = "AUTH.OTP_INVALID"
	LanguageAuthPhoneAlreadyRegistered = "AUTH.PHONE_ALREADY_REGISTERED"

	// user
	LanguageUserNotFound          = "USER.NOT_FOUND"
	LanguageUserEmailAlreadyExist = "USER.EMAIL_ALREADY_EXIST"
	LanguageUserPhoneAlreadyExist = "USER.PHONE_ALREADY_EXIST"
	LanguageUserInvalidPassword   = "USER.INVALID_PASSWORD"
	LanguageUserRoleNotFound      = "USER.ROLE_NOT_FOUND"
	LanguageUserRoleAlreadyExist  = "USER.ROLE_ALREADY_EXIST"

	// product
	LanguageProductNotFound                = "PRODUCT.NOT_FOUND"
	LanguageProductInvalidStatusTransition = "PRODUCT.INVALID_STATUS_TRANSITION"

	// role request
	LanguageRoleRequestAlreadyHaveRole    = "ROLE_REQUEST.ALREADY_HAVE_ROLE"
	LanguageRoleRequestPendingExists      = "ROLE_REQUEST.PENDING_EXISTS"
	LanguageRoleRequestPrerequisiteNotMet = "ROLE_REQUEST.PREREQUISITE_NOT_MET"
	LanguageRoleRequestMissingBidderInfo  = "ROLE_REQUEST.MISSING_BIDDER_INFO"
	LanguageRoleRequestMissingSellerInfo  = "ROLE_REQUEST.MISSING_SELLER_INFO"
	LanguageRoleRequestNotFound           = "ROLE_REQUEST.NOT_FOUND"

	// withdrawal request
	LanguageWithdrawalRequestInsufficientBalance = "WITHDRAWAL_REQUEST.INSUFFICIENT_BALANCE"
	LanguageWithdrawalRequestNoBankAccount       = "WITHDRAWAL_REQUEST.NO_BANK_ACCOUNT"
	LanguageWithdrawalRequestNotFound            = "WITHDRAWAL_REQUEST.NOT_FOUND"

	// auction
	LanguageAuctionNotFound           = "AUCTION.NOT_FOUND"
	LanguageAuctionProductNotVerified = "AUCTION.PRODUCT_NOT_VERIFIED"
	LanguageAuctionNotScheduled       = "AUCTION.NOT_SCHEDULED"
	LanguageAuctionInvalidTimeRange   = "AUCTION.INVALID_TIME_RANGE"

	// bid
	LanguageBidNotFound = "BID.NOT_FOUND"

	// system
	LanguageSystemUnauthorized          = "SYSTEM.UNAUTHORIZED"
	LanguageSystemForbidden             = "SYSTEM.FORBIDDEN"
	LanguageSystemInvalidRequestPayload = "SYSTEM.INVALID_REQUEST_PAYLOAD"
	LanguageSystemInternalServerError   = "SYSTEM.INTERNAL_SERVER_ERROR"
	LanguageSystemNotFound              = "SYSTEM.NOT_FOUND"
	LanguageSystemMustBeAValidUuid      = "SYSTEM.MUST_BE_A_VALID_UUID"

	// file
	LanguageFileExtensionIsNotSupported = "FILE.EXTENSION_IS_NOT_SUPPORTED"
	LanguageFileSizeIs0B                = "FILE.SIZE_IS_0B"
	LanguageFileFileNotExist            = "FILE.FILE_NOT_EXIST"
	LanguageFileSomeFileNotExist        = "FILE.SOME_FILE_NOT_EXIST"
	LanguageFileMaximumFileSizeIsXMB    = "FILE.MAXIMUM_FILE_SIZE_IS_X_MB"
)
