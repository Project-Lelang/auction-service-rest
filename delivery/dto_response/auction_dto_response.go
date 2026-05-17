package dto_response

import (
	"context"

	"auction-service/data_type"
	"auction-service/model"
	"auction-service/util"
)

// AuctionResponse represents a single auction in API responses.
type AuctionResponse struct {
	Id            string             `json:"id"             example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductId     string             `json:"product_id"     example:"550e8400-e29b-41d4-a716-446655440001"`
	StartingPrice float64            `json:"starting_price" example:"100000"`
	StartTime     data_type.DateTime `json:"start_time"`
	EndTime       data_type.DateTime `json:"end_time"`
	Status        string             `json:"status"         example:"SCHEDULED"`
	Fee           float64            `json:"fee"            example:"5000"`
	Product       *ProductResponse   `json:"product,omitempty"`
	Timestamp
} // @name AuctionResponse

func NewAuctionResponse(ctx context.Context, a model.Auction) AuctionResponse {
	r := AuctionResponse{
		Id:            a.Id,
		ProductId:     a.ProductId,
		StartingPrice: a.StartingPrice,
		StartTime:     a.StartTime,
		EndTime:       a.EndTime,
		Status:        a.Status,
		Fee:           a.Fee,
		Timestamp:     Timestamp(a.Timestamp),
	}

	if a.Product != nil {
		r.Product = util.Pointer(NewProductResponse(ctx, *a.Product))
	}

	return r
}
