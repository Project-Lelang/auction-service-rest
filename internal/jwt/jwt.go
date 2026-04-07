package jwt

import (
	"errors"

	jwtLib "github.com/golang-jwt/jwt/v4"
)

var ErrInvalidToken = errors.New("invalid token")

type jwt struct {
	secretKey []byte
}

func (j *jwt) signMethod() jwtLib.SigningMethod {
	return jwtLib.SigningMethodHS256
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

func (j *jwt) tokenType() string {
	return "Bearer"
}

func (j *jwt) Generate(payload Payload) (*Token, error) {
	token := jwtLib.NewWithClaims(j.signMethod(), jwtLib.RegisteredClaims{
		Audience:  []string{payload.UserId},
		ExpiresAt: &jwtLib.NumericDate{Time: payload.ExpiredAt},
		ID:        payload.UserAccessTokenId,
		IssuedAt:  &jwtLib.NumericDate{Time: payload.CreatedAt},
		NotBefore: &jwtLib.NumericDate{Time: payload.CreatedAt},
		Subject:   payload.UserId,
	})

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

	claims := jwtLib.RegisteredClaims{}
	_, err = jwtLib.ParseWithClaims(signedToken, &claims, func(t *jwtLib.Token) (interface{}, error) {
		if t.Method != j.signMethod() {
			return nil, ErrInvalidToken
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	payload := Payload{
		UserAccessTokenId: claims.ID,
		UserId:            claims.Subject,
		CreatedAt:         claims.IssuedAt.Time,
		ExpiredAt:         claims.ExpiresAt.Time,
	}

	return &payload, nil
}

func NewJwt(secretKey []byte) Jwt {
	return &jwt{
		secretKey: secretKey,
	}
}
