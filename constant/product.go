package constant

// ProductStatus values represent the lifecycle of a product listing.
const (
	ProductStatusDraft     = "DRAFT"
	ProductStatusRequest   = "REQUEST"
	ProductStatusVerified  = "VERIFIED"
	ProductStatusRejected  = "REJECTED"
	ProductStatusOnBids    = "ON_BIDS"
	ProductStatusCompleted = "COMPLETED"
)

// ProductCondition values.
const (
	ProductConditionNew      = "NEW"
	ProductConditionPreloved = "PRELOVED"
)

// validProductStatusTransitions defines which status changes are allowed.
// Key = current status, Value = set of permitted next statuses.
var ValidProductStatusTransitions = map[string][]string{
	ProductStatusDraft:    {ProductStatusRequest},
	ProductStatusRequest:  {ProductStatusVerified, ProductStatusRejected},
	ProductStatusVerified: {ProductStatusOnBids},
	ProductStatusOnBids:   {ProductStatusCompleted},
	ProductStatusRejected: {ProductStatusDraft},
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
