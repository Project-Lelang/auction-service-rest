package model

import "time"

type Token struct {
	AccessToken           string
	AccessTokenExpiredAt  time.Time
	RefreshToken          string
	RefreshTokenExpiredAt time.Time
	TokenType             string
}
