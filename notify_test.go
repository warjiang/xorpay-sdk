package xorpay

import (
	"net/url"
	"testing"
)

func TestVerifyNotify(t *testing.T) {
	client, err := NewClient(Config{AppID: "704046", AppSecret: "8d15136f11f3458a91dfe84a0145c612"})
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	payload := NotifyPayload{
		AOID:     "dfa20e231xf64069a615ecfee0527726",
		OrderID:  "order-1",
		PayPrice: "50.00",
		PayTime:  "1714694422",
	}
	payload.Sign = SignNotify(payload.AOID, payload.OrderID, payload.PayPrice, payload.PayTime, "8d15136f11f3458a91dfe84a0145c612")

	if err := client.VerifyNotify(payload); err != nil {
		t.Fatalf("VerifyNotify error: %v", err)
	}

	payload.Sign = "bad-sign"
	if err := client.VerifyNotify(payload); err == nil {
		t.Fatal("expected sign mismatch error")
	}
}

func TestNotifyPayloadFromValues(t *testing.T) {
	v := url.Values{}
	v.Set("aoid", "a1")
	v.Set("order_id", "o1")
	v.Set("pay_price", "1.00")
	v.Set("pay_time", "1714694422")
	v.Set("sign", "s1")

	p := NotifyPayloadFromValues(v)
	if p.AOID != "a1" || p.OrderID != "o1" || p.PayPrice != "1.00" || p.PayTime != "1714694422" || p.Sign != "s1" {
		t.Fatalf("unexpected payload: %+v", p)
	}
}
