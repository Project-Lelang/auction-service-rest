package use_case

import (
	"context"
	"fmt"
	"log"
	"time"

	"auction-service/infrastructure"
	"auction-service/internal/notification"
	"auction-service/util"
)

type NotificationPublisher interface {
	Publish(ctx context.Context, payload notification.Payload) error
}

func publishAuctionNotification(ctx context.Context, publisher NotificationPublisher, payload notification.Payload) {
	if publisher == nil {
		return
	}
	if payload.EventId == "" {
		payload.EventId = util.NewUuid()
	}
	if payload.Timestamp.IsZero() {
		payload.Timestamp = time.Now().UTC()
	}
	if payload.DataPayload == nil {
		payload.DataPayload = map[string]string{}
	}
	if err := publisher.Publish(ctx, payload); err != nil {
		log.Printf("[notification publisher] event=%s user=%d failed: %v", payload.EventType, payload.UserId, err)
	}
}

func auctionURL(auctionId int64) string {
	return fmt.Sprintf("/auction/%d", auctionId)
}

var _ NotificationPublisher = infrastructure.NotificationQueueClient(nil)
