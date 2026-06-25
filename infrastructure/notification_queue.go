package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auction-service/global"
	"auction-service/internal/notification"

	"github.com/redis/go-redis/v9"
)

const (
	defaultNotificationQueueName = "notif_queue:auction_events"
	defaultNotificationDLQName   = "notif_queue:auction_events:dlq"
)

type NotificationQueueClient interface {
	Publish(ctx context.Context, payload notification.Payload) error
	Subscribe(ctx context.Context) (notification.Payload, error)
	MoveToDLQ(ctx context.Context, payload notification.Payload, reason string) error
	Close() error
}

type redisNotificationQueueClient struct {
	client    *redis.Client
	queueName string
	dlqName   string
}

func NewNotificationQueueClient(cfg global.RedisConfig, notificationCfg global.NotificationConfig) NotificationQueueClient {
	queueName := notificationCfg.QueueName
	if queueName == "" {
		queueName = defaultNotificationQueueName
	}
	dlqName := notificationCfg.DLQName
	if dlqName == "" {
		dlqName = defaultNotificationDLQName
	}

	return &redisNotificationQueueClient{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		queueName: queueName,
		dlqName:   dlqName,
	}
}

func (c *redisNotificationQueueClient) Publish(ctx context.Context, payload notification.Payload) error {
	if err := payload.Validate(); err != nil {
		return err
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.client.LPush(ctx, c.queueName, b).Err()
}

func (c *redisNotificationQueueClient) Subscribe(ctx context.Context) (notification.Payload, error) {
	result, err := c.client.BRPop(ctx, 0*time.Second, c.queueName).Result()
	if err != nil {
		return notification.Payload{}, err
	}
	if len(result) != 2 {
		return notification.Payload{}, fmt.Errorf("unexpected redis BRPOP result")
	}

	var payload notification.Payload
	if err := json.Unmarshal([]byte(result[1]), &payload); err != nil {
		return notification.Payload{}, err
	}
	return payload, payload.Validate()
}

func (c *redisNotificationQueueClient) MoveToDLQ(ctx context.Context, payload notification.Payload, reason string) error {
	item := struct {
		Payload  notification.Payload `json:"payload"`
		Reason   string               `json:"reason"`
		FailedAt time.Time            `json:"failed_at"`
	}{
		Payload:  payload,
		Reason:   reason,
		FailedAt: time.Now().UTC(),
	}
	b, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return c.client.LPush(ctx, c.dlqName, b).Err()
}

func (c *redisNotificationQueueClient) Close() error {
	return c.client.Close()
}
