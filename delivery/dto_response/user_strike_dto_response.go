package dto_response

import (
	"auction-service/data_type"
	"auction-service/model"
)

type UserStrikeResponse struct {
	Id           int64                  `json:"id"            example:"1"`
	BidderId     int64                  `json:"bidder_id"     example:"2"`
	AuctionId    int64                  `json:"auction_id"    example:"3"`
	SellerId     int64                  `json:"seller_id"     example:"4"`
	StrikeReason string                 `json:"strike_reason" example:"UNPAID_AUCTION"`
	Status       string                 `json:"status"        example:"ACTIVE"`
	ExpiredAt    data_type.NullDateTime `json:"expired_at"`
	Timestamp
} // @name UserStrikeResponse

func NewUserStrikeResponse(s model.UserStrike) UserStrikeResponse {
	return UserStrikeResponse{
		Id:           s.Id,
		BidderId:     s.BidderId,
		AuctionId:    s.AuctionId,
		SellerId:     s.SellerId,
		StrikeReason: s.StrikeReason,
		Status:       s.Status,
		ExpiredAt:    s.ExpiredAt,
		Timestamp:    Timestamp(s.Timestamp),
	}
}
