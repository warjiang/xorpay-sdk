# xorpay-sdk (Go)

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-blue)](https://go.dev)
[![CI](https://github.com/warjiang/xorpay-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/warjiang/xorpay-sdk/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Go SDK for [XorPay](https://xorpay.com), aligned with the official Java demo flow and extended to full public API coverage.

- **Zero dependencies** — built entirely on the Go standard library (`net/http`).
- **Strongly typed** — request/response structs with validation for every endpoint.
- **Secure by default** — automatic signature generation and callback verification.
- **Well tested** — comprehensive offline, mock-based unit tests with race detection.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Supported Payment Types](#supported-payment-types)
- [Configuration](#configuration)
- [Examples](#examples)
- [Notify Verification](#notify-verification)
- [Error Handling](#error-handling)
- [Documentation](#documentation)
- [Development](#development)
- [License](#license)

## Features

| API | SDK Method | Endpoint |
|-----|-----------|----------|
| Native / JSAPI / Mini Program Pay | `CreatePay` | `POST /api/pay/{aid}` |
| Cashier Pay | `CreateCashier` | `POST /api/cashier/{aid}` |
| Barcode Pay | `CreateBarcodePay` | `POST /api/barcode_pay/{aid}` |
| Query by AOID | `QueryByAOID` | `GET /api/query/{aoid}` |
| Query by Order ID | `QueryByOrderID` | `GET /api/query2/{aid}` |
| Refund | `Refund` | `POST /api/refund/{aoid}` |
| Build OpenID URL | `BuildOpenIDURL` | `/api/openid/{aid}` |
| Build QR URL | `BuildQRURL` | `/qr?data=...` |
| Notify Verification | `VerifyNotify` | — |

## Installation

Requires **Go 1.25** or later.

```bash
go get github.com/warjiang/xorpay-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/warjiang/xorpay-sdk"
)

func main() {
    client, err := xorpay.NewClient(xorpay.ConfigFromEnv())
    if err != nil {
        panic(err)
    }

    resp, err := client.CreatePay(context.Background(), xorpay.PayRequest{
        Name:      "内容订阅一年期",
        PayType:   "native",
        Price:     "50.00",
        OrderID:   "demo-0001",
        NotifyURL: "https://merchant.example.com/xorpay_notify",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(resp.Status, resp.AOID, string(resp.Info))
}
```

Or pass credentials explicitly:

```go
client, err := xorpay.NewClient(xorpay.Config{
    AppID:     "mock_appid",
    AppSecret: "mock_secret",
})
```

## Supported Payment Types

| `PayType` | Description | SDK Method | Notes |
|-----------|-------------|-----------|-------|
| `native` | Native / QR code payment | `CreatePay` | Returns a payment URL or QR data |
| `cashier` | Cashier page redirect | `CreateCashier` | Use for web checkout pages |
| `jsapi` | WeChat JSAPI | `CreatePay` | Requires `OpenID` |
| `barcode` | Barcode / scan-to-pay | `CreateBarcodePay` | Requires `Barcode` |
| `miniprogram` | WeChat Mini Program | `CreatePay` | Set `IsMini: true` and provide `AppID` |

## Configuration

Load from environment variables:

```go
cfg := xorpay.ConfigFromEnv()
client, err := xorpay.NewClient(cfg)
```

Or build `Config` explicitly:

| Field | Required | Env Var Fallback | Description |
|-------|----------|------------------|-------------|
| `AppID` | yes | `XORPAY_APP_ID` | Your XorPay `aid` |
| `AppSecret` | yes | `XORPAY_APP_SECRET` | Your app secret |
| `BaseURL` | no | — | API base URL (default: `https://xorpay.com`) |
| `NotifyURL` | no | `XORPAY_NOTIFY_URL` | Default notify URL for payment APIs |
| `ReturnURL` | no | `XORPAY_RETURN_URL` | Default return URL for cashier / jsapi |

Client options:

```go
client, err := xorpay.NewClient(cfg,
    xorpay.WithBaseURL("https://xorpay.com"),
    xorpay.WithHTTPClient(customHTTPClient),
    xorpay.WithUserAgent("my-app/1.0"),
)
```

## Examples

Set environment variables first:

```bash
export XORPAY_APP_ID=mock_appid
export XORPAY_APP_SECRET=mock_secret
export XORPAY_NOTIFY_URL=https://merchant.example.com/xorpay_notify
export XORPAY_RETURN_URL=https://merchant.example.com/xorpay_return
```

Run standalone examples:

```bash
go run ./examples/pay_native
go run ./examples/pay_cashier
go run ./examples/pay_jsapi
go run ./examples/pay_barcode
go run ./examples/query
go run ./examples/refund
go run ./examples/notify_verify
go run ./examples/internalcfg
```

### Full Gin Demo

A complete web demo with order state management is available under `examples/gin_app`:

```bash
cd examples/gin_app
go run .
```

Environment variables for the Gin demo:

| Variable | Required | Default |
|----------|----------|---------|
| `XORPAY_APP_ID` | yes | — |
| `XORPAY_APP_SECRET` | yes | — |
| `XORPAY_NOTIFY_URL` | yes | — |
| `XORPAY_RETURN_URL` | no | — |
| `XORPAY_BASE_URL` | no | `https://xorpay.com` |
| `GIN_ADDR` | no | `:8080` |

### Multi-language Demo Inputs

- Java: `./demos/java-demo.zip`
- Python: `./demos/native.py.zip`
- Node.js: `./demos/native.js.zip`
- H5: `./demos/h5.zip`
- H5 Cashier: `./demos/h5-cashier.zip`

## Notify Verification

Always verify the callback signature **before** processing business logic:

```go
func notifyHandler(w http.ResponseWriter, r *http.Request) {
    _ = r.ParseForm()
    payload := xorpay.NotifyPayloadFromValues(r.Form)

    if err := client.VerifyNotify(payload); err != nil {
        http.Error(w, "bad sign", http.StatusBadRequest)
        return
    }

    // Production checklist:
    // 1. Query the database by payload.OrderID (or payload.AOID).
    // 2. If the order is already marked as paid, return "ok" immediately (idempotency).
    // 3. Otherwise, update the order status in a transaction and return "ok" only on success.
    // 4. Do NOT process business logic before signature verification.

    _, _ = w.Write([]byte("ok"))
}
```

Signature rules used by the SDK (consistent with XorPay official docs):

| API | Signature Formula |
|-----|-------------------|
| Pay / Cashier | `name + pay_type + price + order_id + notify_url + app_secret` |
| Barcode Pay | `name + pay_type + price + order_id + notify_url + barcode + app_secret` |
| Query by Order ID | `order_id + app_secret` |
| Refund | `price + app_secret` |
| Notify Verify | `aoid + order_id + pay_price + pay_time + app_secret` |

All signatures are lowercase MD5 over plain concatenated values.

## Error Handling

Business-level non-success responses return `*xorpay.APIError`:

```go
resp, err := client.CreatePay(ctx, req)
if err != nil {
    if xorpay.IsStatus(err, "sign_error") {
        // handle sign mismatch
    }

    var apiErr *xorpay.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.Endpoint, apiErr.Status, apiErr.Message)
    }
}
```

HTTP errors (status >= 400) are also wrapped as `*APIError` with `HTTPStatus` and `RawBody` populated.

## Documentation

- [Getting Started](docs/getting-started.md) — client setup, typical flow, and best practices
- [API Reference](docs/api-reference.md) — detailed method signatures and field descriptions
- [Gin Integration](docs/gin-integration.md) — full web integration guide

## Development

```bash
# Run tests with race detection and coverage
go test -race -coverprofile=coverage.out ./...

# Static analysis
go vet ./...

# Format code
gofmt -w .
```

CI is configured via [GitHub Actions](.github/workflows/ci.yml) and runs on every push to `main` and on pull requests.

## License

[MIT](LICENSE)
