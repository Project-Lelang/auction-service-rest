package jwt

import (
	"strings"

	"auction-service/data_type"
)

type Token struct {
	AccessToken string
	TokenType   string
	ExpiredAt   data_type.DateTime
}

func parseToken(finalizedToken string) (*Token, error) {
	temp := strings.Split(finalizedToken, " ")
	if len(temp) != 2 {
		return nil, ErrInvalidToken
	}

	return &Token{
		AccessToken: temp[1],
		TokenType:   temp[0],
	}, nil
}
