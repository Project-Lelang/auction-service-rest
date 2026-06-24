package infrastructure

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"auction-service/global"
	"auction-service/internal/notification"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type PushSendResult struct {
	MessageId    string
	Token        string
	InvalidToken bool
	Error        error
}

type PushBatchResult struct {
	Responses []PushSendResult
}

type PushClient interface {
	SendMulticast(ctx context.Context, tokens []string, payload notification.Payload) (PushBatchResult, error)
}

type noopPushClient struct{}

func NewNoopPushClient() PushClient {
	return noopPushClient{}
}

func (noopPushClient) SendMulticast(_ context.Context, tokens []string, _ notification.Payload) (PushBatchResult, error) {
	result := PushBatchResult{Responses: make([]PushSendResult, 0, len(tokens))}
	for _, token := range tokens {
		result.Responses = append(result.Responses, PushSendResult{MessageId: "noop", Token: token})
	}
	return result, nil
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

func (c *firebasePushClient) SendMulticast(ctx context.Context, tokens []string, payload notification.Payload) (PushBatchResult, error) {
	if len(tokens) == 0 {
		return PushBatchResult{}, nil
	}

	result := PushBatchResult{Responses: make([]PushSendResult, 0, len(tokens))}
	for start := 0; start < len(tokens); start += 500 {
		end := start + 500
		if end > len(tokens) {
			end = len(tokens)
		}
		batchTokens := tokens[start:end]
		message := &messaging.MulticastMessage{
			Tokens: batchTokens,
			Notification: &messaging.Notification{
				Title: payload.Title,
				Body:  payload.Body,
			},
			Data: buildNotificationData(payload),
		}

		if link := message.Data["auction_url"]; isHTTPURL(link) {
			message.Webpush = &messaging.WebpushConfig{
				FCMOptions: &messaging.WebpushFCMOptions{Link: link},
			}
		}

		response, err := c.client.SendEachForMulticast(ctx, message)
		if err != nil {
			return result, err
		}
		for i, sendResponse := range response.Responses {
			sendResult := PushSendResult{Token: batchTokens[i]}
			if sendResponse.Success {
				sendResult.MessageId = sendResponse.MessageID
			} else {
				sendResult.Error = sendResponse.Error
				sendResult.InvalidToken = messaging.IsUnregistered(sendResponse.Error)
			}
			result.Responses = append(result.Responses, sendResult)
		}
	}
	return result, nil
}

func buildNotificationData(payload notification.Payload) map[string]string {
	data := make(map[string]string, len(payload.DataPayload)+4)
	for k, v := range payload.DataPayload {
		data[k] = v
	}
	data["event_id"] = payload.EventId
	data["event_type"] = payload.EventType
	data["auction_id"] = strconv.FormatInt(payload.AuctionId, 10)
	data["role"] = payload.Role
	return data
}

func isHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}
