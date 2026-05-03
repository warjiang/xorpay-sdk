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
		Name:      "内容订阅一年期",
		PayType:   "native",
		Price:     "50.00",
		OrderID:   fmt.Sprintf("demo-native-%d", time.Now().Unix()),
		NotifyURL: "https://merchant.example.com/xorpay_notify",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("status=%s aoid=%s info=%s\n", resp.Status, resp.AOID, string(resp.Info))
}
