package dto_response

import "auction-service/model"

type NotificationResponse struct {
	Id          int64  `json:"id"           example:"1"`
	UserId      int64  `json:"user_id"      example:"2"`
	Title       string `json:"title"        example:"Outbid!"`
	Body        string `json:"body"         example:"Your bid has been outbid."`
	Type        string `json:"type"         example:"OUTBID"`
	ReferenceId *int64 `json:"reference_id,omitempty" example:"3"`
	IsRead      bool   `json:"is_read"      example:"false"`
	Timestamp
} // @name NotificationResponse

func NewNotificationResponse(n model.Notification) NotificationResponse {
	return NotificationResponse{
		Id:          n.Id,
		UserId:      n.UserId,
		Title:       n.Title,
		Body:        n.Body,
		Type:        n.Type,
		ReferenceId: n.ReferenceId,
		IsRead:      n.IsRead,
		Timestamp:   Timestamp(n.Timestamp),
	}
}
