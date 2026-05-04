package main

import (
	"fmt"
	"log"
	"net/http"

	xorpay "github.com/warjiang/xorpay-sdk"
	"github.com/warjiang/xorpay-sdk/examples/internalcfg"
)

func main() {
	client := internalcfg.MustClient()

	http.HandleFunc("/xorpay_notify", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		payload := xorpay.NotifyPayloadFromValues(r.Form)
		if err := client.VerifyNotify(payload); err != nil {
			http.Error(w, "bad sign", http.StatusBadRequest)
			return
		}

		// TODO: process your business order here.
		// Production notes:
		//   - Persist aoid/order_id mapping to a database, not in-memory maps.
		//   - Handle notify callbacks idempotently: check if the order is already
		//     marked paid before updating; return "ok" even if duplicated.
		_, _ = w.Write([]byte("ok"))
	})

	fmt.Println("listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
