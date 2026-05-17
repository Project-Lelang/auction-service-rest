package constant

// AuctionFee is the fixed platform fee applied to every auction.
const AuctionFee float64 = 5000

// AuctionStatus values represent the lifecycle of an auction.
const (
	AuctionStatusScheduled          = "SCHEDULED"
	AuctionStatusOnGoing            = "ON_GOING"
	AuctionStatusWaitingForPayment  = "WAITING_FOR_PAYMENT"
	AuctionStatusWaitingForShipment = "WAITING_FOR_SHIPMENT"
	AuctionStatusShipped            = "SHIPPED"
	AuctionStatusDelivered          = "DELIVERED"
	AuctionStatusCancelled          = "CANCELLED"
	AuctionStatusCompleted          = "COMPLETED"
)
