# API Reference

## Client

```go
client, err := xorpay.NewClient(cfg, opts...)
```

`cfg` fields:
- `AppID` required (falls back to `XORPAY_APP_ID` env var)
- `AppSecret` required (falls back to `XORPAY_APP_SECRET` env var)
- `BaseURL` optional (default: `https://xorpay.com`)
- `NotifyURL` optional (default notify URL for payment APIs)
- `ReturnURL` optional (default return URL for cashier/jsapi)

Load from environment variables:

```go
cfg := xorpay.ConfigFromEnv() // reads XORPAY_APP_ID, XORPAY_APP_SECRET, XORPAY_NOTIFY_URL, XORPAY_RETURN_URL
client, err := xorpay.NewClient(cfg, opts...)
```

Options:
- `WithHTTPClient(doer)`
- `WithBaseURL(baseURL)`
- `WithUserAgent(userAgent)`

## Payment APIs

### CreatePay

```go
resp, err := client.CreatePay(ctx, xorpay.PayRequest{...})
```

Required fields:
- `Name`
- `PayType`
- `Price`
- `OrderID`
- `NotifyURL` (falls back to client-level default if set)

Optional:
- `OrderUID`
- `ReturnURL` (falls back to client-level default if set)
- `OpenID`
- `AppID` (for mini program scene)
- `IsMini`

### CreateCashier

```go
resp, err := client.CreateCashier(ctx, xorpay.CashierRequest{...})
```

Required fields:
- `Name`
- `PayType`
- `Price`
- `OrderID`
- `NotifyURL` (falls back to client-level default if set)

Optional:
- `OrderUID`
- `ReturnURL` (falls back to client-level default if set)

### CreateBarcodePay

```go
resp, err := client.CreateBarcodePay(ctx, xorpay.BarcodePayRequest{...})
```

Required fields:
- `Name`
- `PayType`
- `Price`
- `OrderID`
- `NotifyURL` (falls back to client-level default if set)
- `Barcode`

## Query APIs

### QueryByAOID

```go
resp, err := client.QueryByAOID(ctx, aoid)
```

### QueryByOrderID

```go
resp, err := client.QueryByOrderID(ctx, orderID)
```

## Refund API

```go
resp, err := client.Refund(ctx, aoid, price)
```

- `aoid` required
- `price` required

## Notify Verification

```go
err := client.VerifyNotify(xorpay.NotifyPayload{...})
```

Required payload fields:
- `AOID`
- `OrderID`
- `PayPrice`
- `PayTime`
- `Sign`

## Utility URLs

### OpenID URL

```go
url, err := client.BuildOpenIDURL(callback)
```

### QR URL

```go
url := client.BuildQRURL(data)
```
