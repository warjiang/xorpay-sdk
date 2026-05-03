# Getting Started

## 1. Create client

```go
client, err := xorpay.NewClient(xorpay.Config{
    AppID:     os.Getenv("XORPAY_APP_ID"),
    AppSecret: os.Getenv("XORPAY_APP_SECRET"),
})
```

- `AppID`: your XorPay `aid` (for tests: `704046`)
- `AppSecret`: your app secret (for tests: `8d15136f11f3458a91dfe84a0145c612`)
- Optional: `WithBaseURL(...)`, `WithHTTPClient(...)`, `WithUserAgent(...)`

## 2. Signature rules

The SDK calculates signatures for you. Rules are kept consistent with XorPay docs.

- pay/cashier: `name + pay_type + price + order_id + notify_url + app_secret`
- barcode_pay: `name + pay_type + price + order_id + notify_url + barcode + app_secret`
- query2: `order_id + app_secret`
- refund: `price + app_secret`
- notify verify: `aoid + order_id + pay_price + pay_time + app_secret`

All are lowercase MD5 over plain concatenated values.

## 3. Typical flow

1. Create order (`CreatePay` or `CreateCashier`)
2. Persist `aoid` and merchant `order_id`
3. Receive notify callback and call `VerifyNotify`
4. If needed, query by `aoid` or `order_id`
5. Refund by `aoid`

## 4. Notify callback verification

```go
func notifyHandler(w http.ResponseWriter, r *http.Request) {
    _ = r.ParseForm()
    payload := xorpay.NotifyPayloadFromValues(r.Form)

    if err := client.VerifyNotify(payload); err != nil {
        http.Error(w, "bad sign", http.StatusBadRequest)
        return
    }

    // process business logic
    _, _ = w.Write([]byte("ok"))
}
```

## 5. Error handling

Business-level non-success responses return `*xorpay.APIError` in create/refund APIs.

```go
if xorpay.IsStatus(err, "sign_error") {
    // handle sign mismatch
}
```
