package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/warjiang/xorpay-sdk"
	"github.com/warjiang/xorpay-sdk/examples/internalcfg"
)

func main() {
	client := internalcfg.MustClient()

	resp, err := client.CreateBarcodePay(context.Background(), xorpay.BarcodePayRequest{
		Name:      "门店订单",
		PayType:   "wechat_barcode",
		Price:     "1.00",
		OrderID:   fmt.Sprintf("demo-barcode-%d", time.Now().Unix()),
		NotifyURL: "https://merchant.example.com/xorpay_notify",
		Barcode:   "134657556554411111",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("status=%s aoid=%s detail=%s\n", resp.Status, resp.AOID, string(resp.Detail))
}
