package internalcfg

import (
	"fmt"

	xorpay "github.com/warjiang/xorpay-sdk"
)

func MustClient() *xorpay.Client {
	client, err := xorpay.NewClient(xorpay.ConfigFromEnv())
	if err != nil {
		panic(fmt.Sprintf("new xorpay client failed: %v", err))
	}
	return client
}
