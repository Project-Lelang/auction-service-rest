package constant

// ProductStatus values represent the lifecycle of a product listing.
const (
	ProductStatusDraft                    = "DRAFT"
	ProductStatusRequest                  = "REQUEST"
	ProductStatusVerified                 = "VERIFIED"
	ProductStatusScheduled                = "SCHEDULED"
	ProductStatusRejected                 = "REJECTED"
	ProductStatusOnBids                   = "ON_BIDS"
	ProductStatusWaitingForPayment        = "WAITING_FOR_PAYMENT"
	ProductStatusWaitingForSellerDecision = "WAITING_FOR_SELLER_DECISION"
	ProductStatusWaitingForBidderAddress  = "WAITING_FOR_BIDDER_ADDRESS"
	ProductStatusWaitingForShipment       = "WAITING_FOR_SHIPMENT"
	ProductStatusShipped                  = "SHIPPED"
	ProductStatusSellerFailedToShip       = "SELLER_FAILED_TO_SHIP"
	ProductStatusCompleted                = "COMPLETED"
)

// ProductCondition values.
const (
	ProductConditionNew      = "NEW"
	ProductConditionPreloved = "PRELOVED"
)

// validProductStatusTransitions defines which status changes are allowed.
// Key = current status, Value = set of permitted next statuses.
var ValidProductStatusTransitions = map[string][]string{
	ProductStatusDraft:                    {ProductStatusRequest},
	ProductStatusRequest:                  {ProductStatusVerified, ProductStatusRejected},
	ProductStatusVerified:                 {ProductStatusScheduled},
	ProductStatusScheduled:                {ProductStatusOnBids, ProductStatusVerified},
	ProductStatusOnBids:                   {ProductStatusWaitingForPayment, ProductStatusVerified},
	ProductStatusWaitingForPayment:        {ProductStatusWaitingForBidderAddress, ProductStatusWaitingForSellerDecision, ProductStatusCompleted, ProductStatusVerified},
	ProductStatusWaitingForSellerDecision: {ProductStatusVerified, ProductStatusWaitingForPayment},
	ProductStatusWaitingForBidderAddress:  {ProductStatusWaitingForShipment},
	ProductStatusWaitingForShipment:       {ProductStatusShipped, ProductStatusSellerFailedToShip},
	ProductStatusShipped:                  {ProductStatusCompleted},
	ProductStatusSellerFailedToShip:       {ProductStatusVerified},
	ProductStatusRejected:                 {ProductStatusDraft, ProductStatusRequest},
}

// ValidProductStatusTransitionFor returns true when the transition from current → next is allowed.
func ValidProductStatusTransitionFor(current, next string) bool {
	allowed, ok := ValidProductStatusTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}
