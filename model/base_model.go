package model

import "auction-service/data_type"

type BaseModel interface {
	TableName() string
	ToMap() map[string]interface{}
	GetCreatedAt() data_type.DateTime
	SetCreatedAt(t data_type.DateTime)
	GetUpdatedAt() data_type.DateTime
	SetUpdatedAt(t data_type.DateTime)
}
