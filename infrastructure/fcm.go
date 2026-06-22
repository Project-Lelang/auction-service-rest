package infrastructure

import (
	"context"
	"fmt"
	"strconv"

	"auction-service/global"
	"auction-service/internal/notification"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type PushSendResult struct {
	MessageId    string
	InvalidToken bool
}

type PushClient interface {
	Send(ctx context.Context, token string, payload notification.Payload) (PushSendResult, error)
}

type noopPushClient struct{}

func NewNoopPushClient() PushClient {
	return noopPushClient{}
}

func (noopPushClient) Send(_ context.Context, _ string, _ notification.Payload) (PushSendResult, error) {
	return PushSendResult{MessageId: "noop"}, nil
}

type firebasePushClient struct {
	client *messaging.Client
}

func NewFirebasePushClient(ctx context.Context, cfg *global.FirebaseConfig) (PushClient, error) {
	if cfg == nil || cfg.ServiceAccountPath == "" {
		return NewNoopPushClient(), nil
	}

	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(cfg.ServiceAccountPath))
	if err != nil {
		return nil, fmt.Errorf("firebase app init: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging init: %w", err)
	}
	return &firebasePushClient{client: client}, nil
}

func (c *firebasePushClient) Send(ctx context.Context, token string, payload notification.Payload) (PushSendResult, error) {
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Data: payload.DataPayload,
	}
	if msg.Data == nil {
		msg.Data = map[string]string{}
	}
	msg.Data["event_id"] = payload.EventId
	msg.Data["event_type"] = payload.EventType
	msg.Data["auction_id"] = strconv.FormatInt(payload.AuctionId, 10)
	msg.Data["role"] = payload.Role

	id, err := c.client.Send(ctx, msg)
	if err != nil {
		return PushSendResult{
			InvalidToken: messaging.IsRegistrationTokenNotRegistered(err) || messaging.IsInvalidArgument(err),
		}, err
	}
	return PushSendResult{MessageId: id}, nil
}
