package main

import (
	"context"
	"fmt"
	"log"
	"time"

	xorpay "github.com/warjiang/xorpay-sdk"
	"github.com/warjiang/xorpay-sdk/examples/internalcfg"
)

func main() {
	client := internalcfg.MustClient()

	resp, err := client.CreatePay(context.Background(), xorpay.PayRequest{
		Name:      "JSAPI 演示",
		PayType:   "jsapi",
		Price:     "0.01",
		OrderID:   fmt.Sprintf("demo-jsapi-%d", time.Now().Unix()),
		NotifyURL: "https://merchant.example.com/xorpay_notify",
		ReturnURL: "https://merchant.example.com/pay_return",
		OpenID:    "REPLACE_WITH_XORPAY_OPENID",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("status=%s aoid=%s info=%s\n", resp.Status, resp.AOID, string(resp.Info))
}
