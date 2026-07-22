package notification

import (
	"fmt"
	"strings"
	"time"
)

const (
	RoleBidder = "bidder"
	RoleSeller = "seller"

	EventOutbid                 = "OUTBID"
	EventWinAwaitingPay         = "WIN_AWAITING_PAY"
	EventAuctionEnd             = "AUCTION_END"
	EventAuctionUnpaid          = "AUCTION_UNPAID"
	EventAuctionStart           = "AUCTION_START"
	EventBidderAddressConfirmed = "BIDDER_ADDRESS_CONFIRMED"
	EventBidderAddressExpired   = "BIDDER_ADDRESS_EXPIRED"
	EventShipmentShipped        = "SHIPMENT_SHIPPED"
	EventShipmentDelivered      = "SHIPMENT_DELIVERED"
	EventShipmentCompleted      = "SHIPMENT_COMPLETED"
	EventShipmentRefunded       = "SHIPMENT_REFUNDED"
	EventShipmentAutoCompleted  = "SHIPMENT_AUTO_COMPLETED"
	EventShipmentDeadlineMissed = "SHIPMENT_DEADLINE_MISSED"
)

var eventsByRole = map[string]map[string]struct{}{
	RoleBidder: {
		EventOutbid:                {},
		EventWinAwaitingPay:        {},
		EventBidderAddressExpired:  {},
		EventShipmentShipped:       {},
		EventShipmentDelivered:     {},
		EventShipmentCompleted:     {},
		EventShipmentRefunded:      {},
		EventShipmentAutoCompleted: {},
	},
	RoleSeller: {
		EventAuctionEnd:             {},
		EventAuctionUnpaid:          {},
		EventAuctionStart:           {},
		EventBidderAddressConfirmed: {},
		EventBidderAddressExpired:   {},
		EventShipmentDelivered:      {},
		EventShipmentCompleted:      {},
		EventShipmentAutoCompleted:  {},
		EventShipmentDeadlineMissed: {},
	},
}

type Payload struct {
	EventId     string            `json:"event_id"`
	UserId      int64             `json:"user_id"`
	Role        string            `json:"role"`
	EventType   string            `json:"event_type"`
	AuctionId   int64             `json:"auction_id"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	DataPayload map[string]string `json:"data_payload"`
	Timestamp   time.Time         `json:"timestamp"`
	Attempt     int               `json:"attempt,omitempty"`
	Persisted   bool              `json:"persisted,omitempty"`
}

func (p Payload) Validate() error {
	if strings.TrimSpace(p.EventId) == "" {
		return fmt.Errorf("event_id is required")
	}
	if p.UserId <= 0 {
		return fmt.Errorf("user_id is required")
	}
	if p.AuctionId <= 0 {
		return fmt.Errorf("auction_id is required")
	}
	if strings.TrimSpace(p.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(p.Body) == "" {
		return fmt.Errorf("body is required")
	}
	if p.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	events, ok := eventsByRole[p.Role]
	if !ok {
		return fmt.Errorf("unsupported notification role: %s", p.Role)
	}
	if _, ok := events[p.EventType]; !ok {
		return fmt.Errorf("unsupported notification event_type %s for role %s", p.EventType, p.Role)
	}
	return nil
}
