# API Reference

## Client

```go
client, err := xorpay.NewClient(cfg, opts...)
```

`cfg` fields:
- `AppID` required
- `AppSecret` required
- `BaseURL` optional (default: `https://xorpay.com`)

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
- `NotifyURL`

Optional:
- `OrderUID`
- `ReturnURL`
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
- `NotifyURL`

Optional:
- `OrderUID`
- `ReturnURL`

### CreateBarcodePay

```go
resp, err := client.CreateBarcodePay(ctx, xorpay.BarcodePayRequest{...})
```

Required fields:
- `Name`
- `PayType`
- `Price`
- `OrderID`
- `NotifyURL`
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
