package jwt

import "time"

type Payload struct {
	UserAccessTokenId string
	UserId            string
	CreatedAt         time.Time
	ExpiredAt         time.Time
}
