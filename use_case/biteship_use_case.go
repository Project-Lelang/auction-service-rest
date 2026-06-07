package use_case

import (
	"context"

	"auction-service/delivery/dto_request"
	"auction-service/infrastructure"
)

// BiteshipUseCase exposes Biteship area search for address forms.
type BiteshipUseCase interface {
	SearchAreas(ctx context.Context, request dto_request.BiteshipSearchAreasRequest) []infrastructure.BiteshipArea
}

type biteshipUseCase struct {
	biteshipClient infrastructure.BiteshipClient
}

func NewBiteshipUseCase(biteshipClient infrastructure.BiteshipClient) BiteshipUseCase {
	return &biteshipUseCase{
		biteshipClient: biteshipClient,
	}
}

func (u *biteshipUseCase) SearchAreas(_ context.Context, request dto_request.BiteshipSearchAreasRequest) []infrastructure.BiteshipArea {
	if u.biteshipClient == nil {
		return []infrastructure.BiteshipArea{}
	}
	areas, err := u.biteshipClient.SearchAreas(request.Keyword)
	panicIfErr(err)
	if areas == nil {
		return []infrastructure.BiteshipArea{}
	}
	return areas
}
