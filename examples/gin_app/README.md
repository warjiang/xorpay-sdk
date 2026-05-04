# Gin App Demo

A complete web application demonstrating the XorPay Go SDK with order state management.

## Run

```bash
cd examples/gin_app
go run .
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `XORPAY_APP_ID` | yes | — | XorPay app ID (`aid`) |
| `XORPAY_APP_SECRET` | yes | — | XorPay app secret |
| `XORPAY_NOTIFY_URL` | yes | — | Default notify callback URL |
| `XORPAY_RETURN_URL` | no | — | Default return URL for cashier |
| `XORPAY_BASE_URL` | no | `https://xorpay.com` | XorPay API base URL |
| `GIN_ADDR` | no | `:8080` | HTTP server listen address |

## API Endpoints

- `GET /health` — health check
- `POST /api/pay/native` — create native pay order
- `POST /api/pay/cashier` — create cashier order
- `POST /api/pay/barcode` — create barcode pay order
- `GET /api/orders/:order_id` — query order status
- `POST /api/refunds` — refund order
- `POST /xorpay/notify` — XorPay async notify callback

## Notes

- Order state is stored **in-memory** for demo purposes.
- In production, replace with a database and ensure idempotent notify handling.
