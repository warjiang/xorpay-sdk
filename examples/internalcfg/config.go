package internalcfg

import (
	"fmt"
	"os"

	xorpay "github.com/warjiang/xorpay-sdk"
)

func MustClient() *xorpay.Client {
	appID := os.Getenv("XORPAY_APP_ID")
	secret := os.Getenv("XORPAY_APP_SECRET")
	if appID == "" || secret == "" {
		panic("set XORPAY_APP_ID and XORPAY_APP_SECRET first")
	}

	client, err := xorpay.NewClient(xorpay.Config{
		AppID:     appID,
		AppSecret: secret,
	})
	if err != nil {
		panic(fmt.Sprintf("new xorpay client failed: %v", err))
	}
	return client
}
