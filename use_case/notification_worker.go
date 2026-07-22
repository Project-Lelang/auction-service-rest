package use_case

import (
	"context"
	"fmt"
	"math"
	"time"

	"auction-service/infrastructure"
	"auction-service/internal/notification"
	"auction-service/model"
	"auction-service/repository"
)

type NotificationWorkerConfig struct {
	WorkerCount int
	MaxRetries  int
	RetryBase   time.Duration
}

type NotificationWorker struct {
	cfg               NotificationWorkerConfig
	queue             infrastructure.NotificationQueueClient
	pushClient        infrastructure.PushClient
	repositoryManager repository.RepositoryManager
	logger            infrastructure.LoggerStack
	jobs              chan notification.Payload
}

func NewNotificationWorker(
	cfg NotificationWorkerConfig,
	queue infrastructure.NotificationQueueClient,
	pushClient infrastructure.PushClient,
	repositoryManager repository.RepositoryManager,
	logger infrastructure.LoggerStack,
) *NotificationWorker {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 5
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryBase <= 0 {
		cfg.RetryBase = 500 * time.Millisecond
	}
	return &NotificationWorker{
		cfg:               cfg,
		queue:             queue,
		pushClient:        pushClient,
		repositoryManager: repositoryManager,
		logger:            logger,
		jobs:              make(chan notification.Payload, cfg.WorkerCount*2),
	}
}

func (w *NotificationWorker) Start(ctx context.Context) {
	for i := 0; i < w.cfg.WorkerCount; i++ {
		go w.worker(ctx, i+1)
	}
	go w.subscribe(ctx)
}

func (w *NotificationWorker) subscribe(ctx context.Context) {
	for {
		payload, err := w.queue.Subscribe(ctx)
		if err != nil {
			if ctx.Err() != nil {
				close(w.jobs)
				return
			}
			w.logf("level=error component=notification_worker action=subscribe error=%q", err.Error())
			time.Sleep(time.Second)
			continue
		}

		select {
		case <-ctx.Done():
			close(w.jobs)
			return
		case w.jobs <- payload:
		}
	}
}

func (w *NotificationWorker) worker(ctx context.Context, workerId int) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-w.jobs:
			if !ok {
				return
			}
			if err := w.process(ctx, &payload); err != nil {
				w.handleFailure(ctx, payload, err)
			} else {
				w.logf(
					"level=info component=notification_worker worker_id=%d event_id=%s event_type=%s user_id=%d status=sent",
					workerId,
					payload.EventId,
					payload.EventType,
					payload.UserId,
				)
			}
		}
	}
}

func (w *NotificationWorker) process(ctx context.Context, payload *notification.Payload) error {
	if err := payload.Validate(); err != nil {
		return err
	}

	if !payload.Persisted {
		if err := w.persistNotification(ctx, *payload); err != nil {
			return err
		}
		payload.Persisted = true
	}

	tokens, err := w.repositoryManager.UserFcmTokenRepository().FetchByUserIds(ctx, []int64{payload.UserId})
	if err != nil {
		return err
	}

	fcmTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		fcmTokens = append(fcmTokens, token.FcmToken)
	}

	result, err := w.pushClient.SendMulticast(ctx, fcmTokens, *payload)
	if err != nil {
		return err
	}
	if len(result.Responses) != len(tokens) {
		return fmt.Errorf("fcm response count mismatch: got %d want %d", len(result.Responses), len(tokens))
	}

	var transientErr error
	for i, sendResult := range result.Responses {
		token := tokens[i]
		if sendResult.Error == nil {
			w.logf(
				"level=info component=notification_worker event_id=%s user_id=%d token_id=%d fcm_message_id=%s status=token_sent",
				payload.EventId,
				payload.UserId,
				token.Id,
				sendResult.MessageId,
			)
			continue
		}

		if sendResult.InvalidToken {
			_ = w.repositoryManager.UserFcmTokenRepository().DeleteByToken(ctx, token.FcmToken)
			w.logf(
				"level=warn component=notification_worker event_id=%s user_id=%d token_id=%d status=invalid_token_cleaned error=%q",
				payload.EventId,
				payload.UserId,
				token.Id,
				sendResult.Error.Error(),
			)
			continue
		}

		transientErr = sendResult.Error
		w.logf(
			"level=error component=notification_worker event_id=%s user_id=%d token_id=%d status=token_send_failed error=%q",
			payload.EventId,
			payload.UserId,
			token.Id,
			sendResult.Error.Error(),
		)
	}

	if transientErr != nil {
		return transientErr
	}

	return nil
}

func (w *NotificationWorker) persistNotification(ctx context.Context, payload notification.Payload) error {
	referenceId := payload.AuctionId
	return w.repositoryManager.NotifcationRepository().Insert(ctx, &model.Notification{
		UserId:      payload.UserId,
		Title:       payload.Title,
		Body:        payload.Body,
		Type:        payload.EventType,
		ReferenceId: &referenceId,
		IsRead:      false,
	})
}

func (w *NotificationWorker) handleFailure(ctx context.Context, payload notification.Payload, err error) {
	if payload.Attempt >= w.cfg.MaxRetries {
		if dlqErr := w.queue.MoveToDLQ(ctx, payload, err.Error()); dlqErr != nil {
			w.logf("level=error component=notification_worker event_id=%s status=dlq_failed error=%q", payload.EventId, dlqErr.Error())
			return
		}
		w.logf("level=error component=notification_worker event_id=%s status=dead_lettered error=%q", payload.EventId, err.Error())
		return
	}

	payload.Attempt++
	delay := w.retryDelay(payload.Attempt)
	w.logf(
		"level=warn component=notification_worker event_id=%s status=retry_scheduled attempt=%d delay_ms=%d error=%q",
		payload.EventId,
		payload.Attempt,
		delay.Milliseconds(),
		err.Error(),
	)

	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if pubErr := w.queue.Publish(context.Background(), payload); pubErr != nil {
				w.logf("level=error component=notification_worker event_id=%s status=retry_publish_failed error=%q", payload.EventId, pubErr.Error())
			}
		}
	}()
}

func (w *NotificationWorker) retryDelay(attempt int) time.Duration {
	multiplier := math.Pow(2, float64(attempt-1))
	return time.Duration(multiplier) * w.cfg.RetryBase
}

func (w *NotificationWorker) logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if w.logger != nil {
		w.logger.WriteAll(msg)
	}
}
