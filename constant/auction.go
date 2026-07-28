package constant

// AuctionFeePercent is the platform fee percentage charged on the winning bid.
const AuctionFeePercent float64 = 5

// AuctionStatus values represent the lifecycle of an auction.
const (
	AuctionStatusScheduled                = "SCHEDULED"
	AuctionStatusOnGoing                  = "ON_GOING"
	AuctionStatusWaitingForPayment        = "WAITING_FOR_PAYMENT"
	AuctionStatusWaitingForSellerDecision = "WAITING_FOR_SELLER_DECISION"
	AuctionStatusWaitingForBidderAddress  = "WAITING_FOR_BIDDER_ADDRESS"
	AuctionStatusWaitingForShipment       = "WAITING_FOR_SHIPMENT"
	AuctionStatusShipped                  = "SHIPPED"
	AuctionStatusDelivered                = "DELIVERED"
	AuctionStatusSellerFailedToShip       = "SELLER_FAILED_TO_SHIP"
	AuctionStatusCancelled                = "CANCELLED"
	AuctionStatusCompleted                = "COMPLETED"
)

// AuctionWinner status
const (
	AuctionWinnerStatusOnGoing           = "ON_GOING"
	AuctionWinnerStatusWaitingForPayment = "WAITING_FOR_PAYMENT"
	AuctionWinnerStatusCancelled         = "CANCELLED"
	AuctionWinnerStatusCompleted         = "COMPLETED"
)

// UserStrike reason
const (
	UserStrikeReasonUnpaidAuction     = "UNPAID_AUCTION"
	UserStrikeReasonCancelledBySeller = "CANCELLED_BY_SELLER"
)

// UserStrike status
const (
	UserStrikeStatusActive   = "ACTIVE"
	UserStrikeStatusAppealed = "APPEALED"
	UserStrikeStatusRemoved  = "REMOVED"
)

// SecondChanceOffer status
const (
	SecondChanceOfferStatusPending  = "PENDING"
	SecondChanceOfferStatusAccepted = "ACCEPTED"
	SecondChanceOfferStatusRejected = "REJECTED"
	SecondChanceOfferStatusExpired  = "EXPIRED"
)

// Payment status
const (
	PaymentStatusWaitingForPayment = "WAITING_FOR_PAYMENT"
	PaymentStatusCancelled         = "CANCELLED"
	PaymentStatusCompleted         = "COMPLETED"
	PaymentStatusRefunded          = "REFUNDED"
	PaymentStatusExpired           = "EXPIRED"
	PaymentStatusFailed            = "FAILED"
)

// PaymentMethod type
const (
	PaymentMethodTypeBankTransfer = "BANK_TRANSFER"
	PaymentMethodTypeEWallet      = "E_WALLET"
	PaymentMethodTypeCreditCard   = "CREDIT_CARD"
	PaymentMethodTypeQris         = "QRIS"
	PaymentMethodTypeMidtrans     = "MIDTRANS"
)

// PaymentMethod code
const (
	PaymentMethodCodeMidtrans = "MIDTRANS"
)
