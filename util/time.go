package util

import (
	"time"

	"auction-service/data_type"
)

func CurrentDateTime() data_type.DateTime {
	return data_type.NewDateTime(time.Now())
}
