package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"auction-service/global"

	"github.com/hibiken/asynq"
)

const (
	TypeAuctionStart       = "auction:start"
	TypeAuctionClose       = "auction:close"
	TypePaymentExpiry      = "payment:expire"
	TypeSecondChanceExpiry = "second_chance_offer:expire"
	TypeShipmentAddressDue = "shipment:address_due"
	TypeShipmentShipDue    = "shipment:ship_due"
	TypeShipmentTrackCheck = "shipment:track_check"
	TypeShipmentReceiveDue = "shipment:receive_due"
	taskQueueName          = "default"
	taskRetentionDuration  = 24 * time.Hour
)

// AuctionTaskPayload is the JSON payload carried by auction lifecycle tasks.
type AuctionTaskPayload struct {
	AuctionId int64 `json:"auction_id"`
}

// PaymentTaskPayload is the JSON payload carried by payment expiry tasks.
type PaymentTaskPayload struct {
	PaymentId int64 `json:"payment_id"`
}

// SecondChanceOfferTaskPayload is the JSON payload carried by offer expiry tasks.
type SecondChanceOfferTaskPayload struct {
	OfferId int64 `json:"offer_id"`
}

// ShipmentTaskPayload is the JSON payload carried by shipment deadline tasks.
type ShipmentTaskPayload struct {
	ShipmentId int64 `json:"shipment_id"`
}

// TaskQueueClient schedules and manages auction lifecycle tasks.
type TaskQueueClient interface {
	EnqueueAuctionStart(AuctionId int64, auctionCode string, processAt time.Time) error
	EnqueueAuctionClose(AuctionId int64, auctionCode string, processAt time.Time) error
	EnqueuePaymentExpiry(PaymentId int64, processAt time.Time) error
	EnqueueSecondChanceOfferExpiry(OfferId int64, processAt time.Time) error
	EnqueueShipmentAddressDue(ShipmentId int64, processAt time.Time) error
	EnqueueShipmentShipDue(ShipmentId int64, processAt time.Time) error
	EnqueueShipmentTrackCheck(ShipmentId int64, processAt time.Time) error
	EnqueueShipmentReceiveDue(ShipmentId int64, processAt time.Time) error
	// ReplaceAuctionStart deletes any existing scheduled start task for the
	// auction then re-enqueues it at the new time. Used when OwnUpdate changes
	// the start/end times of a SCHEDULED auction.
	ReplaceAuctionStart(AuctionId int64, auctionCode string, processAt time.Time) error
	// Reset removes every task in the application queue. It is intended only
	// for a fresh database migration, where every queued database ID is stale.
	Reset() error
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
		log.Printf("[task queue] task id conflict for %s; replacing stale task scheduled at %s", taskId, processAt.Format(time.RFC3339))
		if deleteErr := c.inspector.DeleteTask(taskQueueName, taskId); deleteErr != nil &&
			!errors.Is(deleteErr, asynq.ErrTaskNotFound) &&
			!errors.Is(deleteErr, asynq.ErrQueueNotFound) {
			return fmt.Errorf("replace conflicting task %s: %w", taskId, deleteErr)
		}
		_, err = c.client.Enqueue(task)
	}
	if err != nil {
		return err
	}
	return nil
}

func auctionLifecycleTaskID(typeName string, auctionId int64, auctionCode string) string {
	if auctionCode != "" {
		return fmt.Sprintf("%s:%s", typeName, auctionCode)
	}
	return fmt.Sprintf("%s:%d", typeName, auctionId)
}

func (c *asynqTaskQueueClient) EnqueueAuctionStart(auctionId int64, auctionCode string, processAt time.Time) error {
	payload, _ := json.Marshal(AuctionTaskPayload{AuctionId: auctionId})
	return c.enqueue(TypeAuctionStart, auctionLifecycleTaskID(TypeAuctionStart, auctionId, auctionCode), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueAuctionClose(auctionId int64, auctionCode string, processAt time.Time) error {
	payload, _ := json.Marshal(AuctionTaskPayload{AuctionId: auctionId})
	return c.enqueue(TypeAuctionClose, auctionLifecycleTaskID(TypeAuctionClose, auctionId, auctionCode), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueuePaymentExpiry(paymentId int64, processAt time.Time) error {
	payload, _ := json.Marshal(PaymentTaskPayload{PaymentId: paymentId})
	return c.enqueue(TypePaymentExpiry, fmt.Sprintf("%s:%d", TypePaymentExpiry, paymentId), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueSecondChanceOfferExpiry(offerId int64, processAt time.Time) error {
	payload, _ := json.Marshal(SecondChanceOfferTaskPayload{OfferId: offerId})
	return c.enqueue(TypeSecondChanceExpiry, fmt.Sprintf("%s:%d", TypeSecondChanceExpiry, offerId), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueShipmentAddressDue(shipmentId int64, processAt time.Time) error {
	payload, _ := json.Marshal(ShipmentTaskPayload{ShipmentId: shipmentId})
	return c.enqueue(TypeShipmentAddressDue, fmt.Sprintf("%s:%d", TypeShipmentAddressDue, shipmentId), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueShipmentShipDue(shipmentId int64, processAt time.Time) error {
	payload, _ := json.Marshal(ShipmentTaskPayload{ShipmentId: shipmentId})
	return c.enqueue(TypeShipmentShipDue, fmt.Sprintf("%s:%d", TypeShipmentShipDue, shipmentId), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueShipmentTrackCheck(shipmentId int64, processAt time.Time) error {
	payload, _ := json.Marshal(ShipmentTaskPayload{ShipmentId: shipmentId})
	return c.enqueue(TypeShipmentTrackCheck, fmt.Sprintf("%s:%d:%d", TypeShipmentTrackCheck, shipmentId, processAt.Unix()), payload, processAt)
}

func (c *asynqTaskQueueClient) EnqueueShipmentReceiveDue(shipmentId int64, processAt time.Time) error {
	payload, _ := json.Marshal(ShipmentTaskPayload{ShipmentId: shipmentId})
	return c.enqueue(TypeShipmentReceiveDue, fmt.Sprintf("%s:%d", TypeShipmentReceiveDue, shipmentId), payload, processAt)
}

func (c *asynqTaskQueueClient) ReplaceAuctionStart(auctionId int64, auctionCode string, processAt time.Time) error {
	taskID := auctionLifecycleTaskID(TypeAuctionStart, auctionId, auctionCode)
	// Remove the previously scheduled task; ignore "not found" errors.
	_ = c.inspector.DeleteTask(taskQueueName, taskID)
	payload, _ := json.Marshal(AuctionTaskPayload{AuctionId: auctionId})
	return c.enqueue(TypeAuctionStart, taskID, payload, processAt)
}

func (c *asynqTaskQueueClient) Reset() error {
	err := c.inspector.DeleteQueue(taskQueueName, true)
	if errors.Is(err, asynq.ErrQueueNotFound) {
		return nil
	}
	return err
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
