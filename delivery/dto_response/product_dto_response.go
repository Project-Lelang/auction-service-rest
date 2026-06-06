package dto_response

import (
	"context"

	"auction-service/model"
	"auction-service/util"
)

// ProductResponse represents a single product in API responses.
type ProductResponse struct {
	Id              string                         `json:"id"                       example:"550e8400-e29b-41d4-a716-446655440000"`
	UserId          string                         `json:"user_id"                  example:"550e8400-e29b-41d4-a716-446655440001"`
	Name            string                         `json:"name"                     example:"Vintage Camera"`
	Description     *string                        `json:"description,omitempty"    example:"A beautiful vintage camera"`
	Condition       string                         `json:"condition"                example:"NEW"`
	CoverImageLink  *string                        `json:"cover_image_link,omitempty"`
	ImageLinks      []string                       `json:"image_links"`
	WeightGram      int                            `json:"weight_gram"              example:"500"`
	Status          string                         `json:"status"                   example:"DRAFT"`
	User            *UserResponse                  `json:"user,omitempty"`
	StatusHistories []ProductStatusHistoryResponse `json:"status_histories"`
	Timestamp
} // @name ProductResponse

// ProductStatusHistoryResponse represents a single status history entry.
type ProductStatusHistoryResponse struct {
	Id        string  `json:"id"                example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductId string  `json:"product_id"        example:"550e8400-e29b-41d4-a716-446655440001"`
	Status    string  `json:"status"            example:"VERIFIED"`
	Message   *string `json:"message,omitempty" example:"Looks good!"`
	Timestamp
} // @name ProductStatusHistoryResponse

func NewProductResponse(ctx context.Context, p model.Product) ProductResponse {
	r := ProductResponse{
		Id:              p.Id,
		UserId:          p.UserId,
		Name:            p.Name,
		Description:     p.Description,
		Condition:       p.Condition,
		CoverImageLink:  p.CoverImageLink,
		ImageLinks:      p.ImageLinks,
		WeightGram:      p.WeightGram,
		Status:          p.Status,
		StatusHistories: []ProductStatusHistoryResponse{},
		Timestamp:       Timestamp(p.Timestamp),
	}

	if p.User != nil {
		r.User = util.Pointer(NewUserResponse(ctx, *p.User))
	}

	for _, h := range p.StatusHistories {
		r.StatusHistories = append(r.StatusHistories, NewProductStatusHistoryResponse(ctx, h))
	}

	return r
}

func NewProductStatusHistoryResponse(_ context.Context, h model.ProductStatusHistory) ProductStatusHistoryResponse {
	return ProductStatusHistoryResponse{
		Id:        h.Id,
		ProductId: h.ProductId,
		Status:    h.Status,
		Message:   h.Message,
		Timestamp: Timestamp(h.Timestamp),
	}
}
