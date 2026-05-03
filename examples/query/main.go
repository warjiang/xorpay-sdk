package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/warjiang/xorpay-sdk/examples/internalcfg"
)

func main() {
	client := internalcfg.MustClient()
	aoid := os.Getenv("XORPAY_AOID")
	if aoid == "" {
		log.Fatal("set XORPAY_AOID first")
	}

	resp, err := client.QueryByAOID(context.Background(), aoid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("query by aoid: status=%s order_id=%s pay_price=%s\n", resp.Status, resp.OrderID, resp.PayPrice)

	orderID := os.Getenv("XORPAY_ORDER_ID")
	if orderID != "" {
		resp2, err := client.QueryByOrderID(context.Background(), orderID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("query by order_id: status=%s aoid=%s\n", resp2.Status, resp2.AOID)
	}
}
