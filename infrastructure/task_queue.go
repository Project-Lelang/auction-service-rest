package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"auction-service/global"

	"github.com/hibiken/asynq"
)

const (
	TypeAuctionStart      = "auction:start"
	TypeAuctionClose      = "auction:close"
	TypePaymentExpiry     = "payment:expire"
	taskQueueName         = "default"
	taskRetentionDuration = 24 * time.Hour
)

// AuctionTaskPayload is the JSON payload carried by auction lifecycle tasks.
type AuctionTaskPayload struct {
	AuctionId int64 `json:"auction_id"`
}

// PaymentTaskPayload is the JSON payload carried by payment expiry tasks.
type PaymentTaskPayload struct {
	PaymentId int64 `json:"payment_id"`
}

// TaskQueueClient schedules and manages auction lifecycle tasks.
type TaskQueueClient interface {
	EnqueueAuctionStart(AuctionId int64, processAt time.Time) error
	EnqueueAuctionClose(AuctionId int64, processAt time.Time) error
	EnqueuePaymentExpiry(PaymentId int64, processAt time.Time) error
	// ReplaceAuctionStart deletes any existing scheduled start task for the
	// auction then re-enqueues it at the new time. Used when OwnUpdate changes
	// the start/end times of a SCHEDULED auction.
	ReplaceAuctionStart(AuctionId int64, processAt time.Time) error
	Close() error
}

type asynqTaskQueueClient struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

func newRedisClientOpt(cfg global.RedisConfig) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
}

// NewTaskQueueClient creates an asynq-backed task queue client.
func NewTaskQueueClient(cfg global.RedisConfig) TaskQueueClient {
	opt := newRedisClientOpt(cfg)
	return &asynqTaskQueueClient{
		client:    asynq.NewClient(opt),
		inspector: asynq.NewInspector(opt),
	}
}

func (c *asynqTaskQueueClient) enqueue(typeName, taskId string, payload []byte, processAt time.Time) error {
	task := asynq.NewTask(
		typeName,
		payload,
		asynq.TaskID(taskId),
		asynq.ProcessAt(processAt),
		asynq.Queue(taskQueueName),
		asynq.Retention(taskRetentionDuration),
	)
	_, err := c.client.Enqueue(task)
	if err != nil && errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil // task already queued — idempotent
	}
	return err
}

func (c *asynqTaskQueueClient) EnqueueAuctionStart(auctionId int64, processAt time.Time) error {
	payload, _ := json.Marshal(AuctionTaskPayload{AuctionId: auctionId})
	return c.enqueue(TypeAuctionStart, fmt.Sprintf("%s:%d", TypeAuctionStart, auctionId), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueAuctionClose(auctionId int64, processAt time.Time) error {
	payload, _ := json.Marshal(AuctionTaskPayload{AuctionId: auctionId})
	return c.enqueue(TypeAuctionClose, fmt.Sprintf("%s:%d", TypeAuctionClose, auctionId), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueuePaymentExpiry(paymentId int64, processAt time.Time) error {
	payload, _ := json.Marshal(PaymentTaskPayload{PaymentId: paymentId})
	return c.enqueue(TypePaymentExpiry, fmt.Sprintf("%s:%d", TypePaymentExpiry, paymentId), payload, processAt)
}

func (c *asynqTaskQueueClient) ReplaceAuctionStart(auctionId int64, processAt time.Time) error {
	taskID := fmt.Sprintf("%s:%d", TypeAuctionStart, auctionId)
	// Remove the previously scheduled task; ignore "not found" errors.
	_ = c.inspector.DeleteTask(taskQueueName, taskID)
	payload, _ := json.Marshal(AuctionTaskPayload{AuctionId: auctionId})
	return c.enqueue(TypeAuctionStart, taskID, payload, processAt)
}

func (c *asynqTaskQueueClient) Close() error {
	_ = c.inspector.Close()
	return c.client.Close()
}

// NewAsynqServer creates an asynq server with 500 ms polling for near-zero scheduling delay.
func NewAsynqServer(cfg global.RedisConfig) *asynq.Server {
	opt := newRedisClientOpt(cfg)
	return asynq.NewServer(opt, asynq.Config{
		Concurrency:              10,
		DelayedTaskCheckInterval: 500 * time.Millisecond, // check scheduled→pending every 500ms
		Queues: map[string]int{
			taskQueueName: 1,
		},
	})
}
