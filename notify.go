package xorpay

import "net/url"

func NotifyPayloadFromValues(v url.Values) NotifyPayload {
	return NotifyPayload{
		AOID:     v.Get("aoid"),
		OrderID:  v.Get("order_id"),
		PayPrice: v.Get("pay_price"),
		PayTime:  v.Get("pay_time"),
		Sign:     v.Get("sign"),
	}
}
