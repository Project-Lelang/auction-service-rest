package constant

// AuctionFee is the fixed platform fee applied to every auction.
const AuctionFee float64 = 5000

// AuctionStatus values represent the lifecycle of an auction.
const (
	AuctionStatusScheduled                = "SCHEDULED"
	AuctionStatusOnGoing                  = "ON_GOING"
	AuctionStatusWaitingForPayment        = "WAITING_FOR_PAYMENT"
	AuctionStatusWaitingForSellerDecision = "WAITING_FOR_SELLER_DECISION"
	AuctionStatusWaitingForBuyerAddress   = "WAITING_FOR_BUYER_ADDRESS"
	AuctionStatusWaitingForShipment       = "WAITING_FOR_SHIPMENT"
	AuctionStatusShipped                  = "SHIPPED"
	AuctionStatusDelivered                = "DELIVERED"
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

// Payment status
const (
	PaymentStatusWaitingForPayment = "WAITING_FOR_PAYMENT"
	PaymentStatusCancelled         = "CANCELLED"
	PaymentStatusCompleted         = "COMPLETED"
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
