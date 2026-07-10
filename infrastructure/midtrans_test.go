package infrastructure

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"
)

func TestMidtransValidateNotification(t *testing.T) {
	serverKey := "Mid-server-secret"
	notification := MidtransNotification{
		OrderId:           "PAY-123",
		StatusCode:        "200",
		GrossAmount:       "150000.00",
		TransactionStatus: "settlement",
	}

	sum := sha512.Sum512([]byte(notification.OrderId + notification.StatusCode + notification.GrossAmount + serverKey))
	notification.SignatureKey = hex.EncodeToString(sum[:])

	client := NewMidtransClient(serverKey, true)
	if !client.ValidateNotification(notification) {
		t.Fatal("expected valid Midtrans notification signature")
	}

	notification.SignatureKey = "invalid"
	if client.ValidateNotification(notification) {
		t.Fatal("expected invalid Midtrans notification signature to be rejected")
	}
}

func TestMidtransValidateNotificationRequiresSignatureFields(t *testing.T) {
	client := NewMidtransClient("Mid-server-secret", true)
	if client.ValidateNotification(MidtransNotification{
		OrderId:     "PAY-123",
		StatusCode:  "200",
		GrossAmount: "150000.00",
	}) {
		t.Fatal("expected notification without signature_key to be rejected")
	}
}

func TestMidtransValidateNotificationRejectsMismatchedSignature(t *testing.T) {
	client := NewMidtransClient("midtrans-test-server-key", true)
	notification := MidtransNotification{
		OrderId:           "PAY-319174df-6379-436b-ba1d-ea8d7b3c9167",
		StatusCode:        "200",
		GrossAmount:       "15.00",
		TransactionStatus: "capture",
		FraudStatus:       "accept",
		SignatureKey:      "16d6f84b2fb0468e2a9cf99a8ac4e5d803d42180347aaa70cb2a7abb13b5c6130458ca9c71956a962c0827637cd3bc7d40b21a8ae9fab12c7c3efe351b18d00a",
	}

	if client.ValidateNotification(notification) {
		t.Fatal("expected mismatched Midtrans notification signature to be rejected")
	}
}
