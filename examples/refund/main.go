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

	resp, err := client.Refund(context.Background(), aoid, "0.01")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("refund status=%s detail=%s\n", resp.Status, string(resp.Detail))
}
