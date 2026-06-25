package constant

// ProductStatus values represent the lifecycle of a product listing.
const (
	ProductStatusDraft                    = "DRAFT"
	ProductStatusRequest                  = "REQUEST"
	ProductStatusVerified                 = "VERIFIED"
	ProductStatusRejected                 = "REJECTED"
	ProductStatusOnBids                   = "ON_BIDS"
	ProductStatusWaitingForPayment        = "WAITING_FOR_PAYMENT"
	ProductStatusWaitingForSellerDecision = "WAITING_FOR_SELLER_DECISION"
	ProductStatusWaitingForBuyerAddress   = "WAITING_FOR_BUYER_ADDRESS"
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
	ProductStatusVerified:                 {ProductStatusOnBids},
	ProductStatusOnBids:                   {ProductStatusWaitingForPayment, ProductStatusVerified},
	ProductStatusWaitingForPayment:        {ProductStatusWaitingForBuyerAddress, ProductStatusWaitingForSellerDecision, ProductStatusCompleted, ProductStatusVerified},
	ProductStatusWaitingForSellerDecision: {ProductStatusVerified, ProductStatusWaitingForPayment},
	ProductStatusWaitingForBuyerAddress:   {ProductStatusWaitingForShipment},
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
