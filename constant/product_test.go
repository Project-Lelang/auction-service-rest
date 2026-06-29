package constant

import "testing"

func TestScheduledProductStatusTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		valid   bool
	}{
		{name: "verified product can be scheduled", current: ProductStatusVerified, next: ProductStatusScheduled, valid: true},
		{name: "scheduled product can start bidding", current: ProductStatusScheduled, next: ProductStatusOnBids, valid: true},
		{name: "scheduled product can return to verified", current: ProductStatusScheduled, next: ProductStatusVerified, valid: true},
		{name: "verified product cannot skip scheduled", current: ProductStatusVerified, next: ProductStatusOnBids, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidProductStatusTransitionFor(tt.current, tt.next); got != tt.valid {
				t.Fatalf("ValidProductStatusTransitionFor(%q, %q) = %v, want %v", tt.current, tt.next, got, tt.valid)
			}
		})
	}
}
