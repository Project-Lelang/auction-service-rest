package use_case

import (
	"regexp"
	"testing"
)

func TestNewPaymentCode(t *testing.T) {
	t.Parallel()

	allowed := regexp.MustCompile(`^PAY-[0-9a-f-]{36}$`)
	seen := make(map[string]struct{}, 100)

	for i := 0; i < 100; i++ {
		code := newPaymentCode()
		if len(code) > 50 {
			t.Fatalf("payment code exceeds Midtrans limit: %q", code)
		}
		if !allowed.MatchString(code) {
			t.Fatalf("payment code contains an unexpected format: %q", code)
		}
		if _, exists := seen[code]; exists {
			t.Fatalf("duplicate payment code generated: %q", code)
		}
		seen[code] = struct{}{}
	}
}
