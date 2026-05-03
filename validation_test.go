package xorpay

import "testing"

func TestRequestValidation(t *testing.T) {
	if err := (PayRequest{}).Validate(); err == nil {
		t.Fatal("expected PayRequest validation error")
	}
	if err := (CashierRequest{}).Validate(); err == nil {
		t.Fatal("expected CashierRequest validation error")
	}
	if err := (BarcodePayRequest{}).Validate(); err == nil {
		t.Fatal("expected BarcodePayRequest validation error")
	}
	if err := (NotifyPayload{}).Validate(); err == nil {
		t.Fatal("expected NotifyPayload validation error")
	}
}

func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("expected NewClient validation error")
	}
}
