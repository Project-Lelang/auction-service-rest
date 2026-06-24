package jwt

import "auction-service/data_type"

type Payload struct {
	Id        int64
	Email     string
	Roles     []string
	CreatedAt data_type.DateTime
	ExpiredAt data_type.DateTime
}
