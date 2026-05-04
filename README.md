# xorpay-sdk (Go)

Go SDK for XorPay, aligned with the Java demo flow and extended to full public API coverage in the docs.

## Features

- Full API wrappers:
  - `CreatePay` -> `/api/pay/{aid}`
  - `CreateCashier` -> `/api/cashier/{aid}`
  - `CreateBarcodePay` -> `/api/barcode_pay/{aid}`
  - `QueryByAOID` -> `/api/query/{aoid}`
  - `QueryByOrderID` -> `/api/query2/{aid}`
  - `Refund` -> `/api/refund/{aoid}`
  - `BuildOpenIDURL` -> `/api/openid/{aid}`
  - `BuildQRURL` -> `/qr?data=...`
  - `VerifyNotify` for callback signature
- Strong typed request/response structures.
- Standard library `net/http` only.
- Comprehensive unit tests (offline, mock-based).

## Install

```bash
go get github.com/warjiang/xorpay-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    xorpay "github.com/warjiang/xorpay-sdk"
)

func main() {
    client, _ := xorpay.NewClient(xorpay.ConfigFromEnv())
    // Or pass credentials explicitly:
    // client, _ := xorpay.NewClient(xorpay.Config{
    //     AppID:     "704046",
    //     AppSecret: "8d15136f11f3458a91dfe84a0145c612",
    // })

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

## Examples

Set env vars first:

```bash
export XORPAY_APP_ID=704046
export XORPAY_APP_SECRET=8d15136f11f3458a91dfe84a0145c612
# optional defaults used by gin_app and ConfigFromEnv()
export XORPAY_NOTIFY_URL=https://merchant.example.com/xorpay_notify
export XORPAY_RETURN_URL=https://merchant.example.com/xorpay_return
```

Run examples:

```bash
go run ./examples/pay_native
go run ./examples/pay_cashier
go run ./examples/pay_jsapi
go run ./examples/pay_barcode
go run ./examples/query
go run ./examples/refund
go run ./examples/notify_verify
```

### Full Gin Demo

A complete web demo with order state management:

```bash
cd examples/gin_app
go run .
```

Environment variables for the gin demo:

| Variable | Required | Default |
|----------|----------|---------|
| `XORPAY_APP_ID` | yes | — |
| `XORPAY_APP_SECRET` | yes | — |
| `XORPAY_NOTIFY_URL` | yes | — |
| `XORPAY_RETURN_URL` | no | — |
| `XORPAY_BASE_URL` | no | `https://xorpay.com` |
| `GIN_ADDR` | no | `:8080` |

## Multi-language Demo Inputs

- Java: `./demos/java-demo.zip`
- Python: `./demos/native.py.zip`
- Node.js: `./demos/native.js.zip`
- H5: `./demos/h5.zip`
- H5 Cashier: `./demos/h5-cashier.zip`

## Documentation

- [Getting Started](docs/getting-started.md)
- [API Reference](docs/api-reference.md)

## Quality

- Unit tests: `go test ./...`
- Vet: `go vet ./...`
- Formatting: `gofmt -w .`
- CI: GitHub Actions in `.github/workflows/ci.yml`
