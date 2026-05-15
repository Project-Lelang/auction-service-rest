package jwt

import (
	"errors"

	"auction-service/data_type"

	jwtLib "github.com/golang-jwt/jwt/v4"
)

var ErrInvalidToken = errors.New("invalid token")

type customClaims struct {
	Id    string   `json:"id"`
	Phone string   `json:"phone"`
	Roles []string `json:"roles"`
	jwtLib.RegisteredClaims
}

type jwt struct {
	secretKey []byte
}

func (j *jwt) signMethod() jwtLib.SigningMethod {
	return jwtLib.SigningMethodHS256
}

func (j *jwt) tokenType() string {
	return "Bearer"
}

func (j *jwt) finalizeToken(signedToken string) *Token {
	return &Token{
		AccessToken: signedToken,
		TokenType:   j.tokenType(),
	}
}

func (j *jwt) parseToken(finalizedToken string) (string, error) {
	token, err := parseToken(finalizedToken)
	if err != nil {
		return "", ErrInvalidToken
	}
	if token.TokenType != j.tokenType() {
		return "", ErrInvalidToken
	}
	return token.AccessToken, nil
}

func (j *jwt) Generate(payload Payload) (*Token, error) {
	claims := &customClaims{
		Id:    payload.Id,
		Phone: payload.Phone,
		Roles: payload.Roles,
		RegisteredClaims: jwtLib.RegisteredClaims{
			ExpiresAt: &jwtLib.NumericDate{Time: payload.ExpiredAt.Time()},
			IssuedAt:  &jwtLib.NumericDate{Time: payload.CreatedAt.Time()},
			NotBefore: &jwtLib.NumericDate{Time: payload.CreatedAt.Time()},
		},
	}

	token := jwtLib.NewWithClaims(j.signMethod(), claims)
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	finalizedToken := j.finalizeToken(signedToken)
	finalizedToken.ExpiredAt = payload.ExpiredAt
	return finalizedToken, nil
}

func (j *jwt) Parse(finalizedToken string) (*Payload, error) {
	signedToken, err := j.parseToken(finalizedToken)
	if err != nil {
		return nil, err
	}

	claims := &customClaims{}
	_, err = jwtLib.ParseWithClaims(signedToken, claims, func(t *jwtLib.Token) (interface{}, error) {
		if t.Method != j.signMethod() {
			return nil, ErrInvalidToken
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	payload := Payload{
		Id:        claims.Id,
		Phone:     claims.Phone,
		Roles:     claims.Roles,
		CreatedAt: data_type.NewDateTime(claims.IssuedAt.Time),
		ExpiredAt: data_type.NewDateTime(claims.ExpiresAt.Time),
	}
	return &payload, nil
}

func NewJwt(secretKey []byte) Jwt {
	return &jwt{secretKey: secretKey}
}
