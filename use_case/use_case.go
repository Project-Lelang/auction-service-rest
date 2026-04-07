package use_case

import (
	"auction-service/constant"
	"auction-service/delivery/dto_response"
)

func panicIfErr(err error, excludedErrs ...error) {
	if err != nil {
		for _, excludedErr := range excludedErrs {
			if err == excludedErr {
				return
			}
		}
		panic(err)
	}
}

func panicIfRepositoryError(err error, errNoDataMessage string) {
	if err != nil {
		if err == constant.ErrNoData {
			panic(dto_response.NewBadRequestErrorResponse(errNoDataMessage))
		}
		panic(err)
	}
}
