# Gin 接入指南

本指南介绍如何基于 `examples/gin_app` 搭建一个完整的 XorPay 支付接入服务。

## 1. 简介

`gin_app` 是一个基于 [Gin](https://gin-gonic.com/) 框架的示例 Web 服务，演示了：

- 创建支付订单（Native、Cashier、Barcode）
- 接收并验证 XorPay 异步回调通知
- 查询订单状态
- 发起退款
- 订单状态管理

> **注意**：示例中的订单状态保存在**内存**中，生产环境请务必替换为数据库，并保证回调处理的幂等性。

## 2. 环境准备

- Go 1.24+
- XorPay 账号（获取 `aid` 和 `app_secret`）
- 可供外网访问的回调地址（用于接收支付结果通知）

## 3. 配置说明

所有配置通过环境变量读取：

| 环境变量 | 是否必填 | 默认值 | 说明 |
|---------|---------|--------|------|
| `XORPAY_APP_ID` | 是 | — | XorPay 应用 ID（`aid`） |
| `XORPAY_APP_SECRET` | 是 | — | XorPay 应用密钥 |
| `XORPAY_NOTIFY_URL` | 是 | — | 全局默认回调通知地址 |
| `XORPAY_RETURN_URL` | 否 | — | 收银台支付后跳转地址 |
| `XORPAY_BASE_URL` | 否 | `https://xorpay.com` | XorPay API 基础地址（测试环境可修改） |
| `GIN_ADDR` | 否 | `:8080` | HTTP 服务监听地址 |

### 回调地址配置说明

`XORPAY_NOTIFY_URL` 必须满足以下条件：

1. **外网可访问**：XorPay 服务器需要能直接访问该地址，不能使用 `localhost`、`127.0.0.1` 或内网 IP
2. **HTTPS 协议**：生产环境必须使用 HTTPS
3. **响应纯文本 `ok`**：收到回调后需返回 HTTP 200 及响应体 `ok`，否则 XorPay 会判定为通知失败并持续重试

**XorPay 后台配置**：

登录 [XorPay 商户后台](https://xorpay.com/) → 应用设置 → 填写「异步回调地址」，该地址即为默认的 `notify_url`。也可在每次创建订单时通过请求参数单独指定，请求级别优先级更高。

**本地开发调试**：

若本地无外网域名，可使用内网穿透工具临时暴露服务：

```bash
# 使用 ngrok 示例
ngrok http 8080
# 获得 https://xxxx.ngrok-free.app → 将其配置为 XORPAY_NOTIFY_URL
```

### 快速配置示例

```bash
export XORPAY_APP_ID="your_aid"
export XORPAY_APP_SECRET="your_secret"
export XORPAY_NOTIFY_URL="https://your-domain.com/xorpay/notify"
export XORPAY_RETURN_URL="https://your-domain.com/pay/success"
export GIN_ADDR=":8080"
```

## 4. 目录结构

```
examples/gin_app/
├── go.mod      # 模块定义，替换指向本地 SDK
├── main.go     # 完整示例代码（路由、业务逻辑、订单管理）
└── README.md   # 简要说明
```

所有业务逻辑集中在 `main.go` 中，便于快速浏览和理解。

## 5. 本地启动

### 5.1 配置环境变量

在终端中导出必要的环境变量：

```bash
export XORPAY_APP_ID="your_aid"
export XORPAY_APP_SECRET="your_secret"
export XORPAY_NOTIFY_URL="https://your-domain.com/xorpay/notify"
export XORPAY_RETURN_URL="https://your-domain.com/pay/success"
export GIN_ADDR=":8080"
```

> 也可将上述内容写入 `.env` 文件，通过 `export $(cat .env | xargs)` 批量加载。

### 5.2 启动服务

```bash
cd examples/gin_app
go mod tidy
go run .
```

正常启动后控制台将输出：

```
listen on :8080
[GIN-debug] GET    /health                   --> main.(*appServer).router.func1 (3 handlers)
...
```

### 5.3 验证启动

```bash
curl http://localhost:8080/health
```

若返回 `{"status":"ok"}`，说明服务已成功启动。

## 6. 接口详情

### 6.1 健康检查

- **方法**：`GET`
- **路径**：`/health`
- **说明**：服务健康检查

**请求示例**：

```bash
curl http://localhost:8080/health
```

**响应示例**：

```json
{"status":"ok"}
```

### 6.2 创建 Native 支付

- **方法**：`POST`
- **路径**：`/api/pay/native`
- **说明**：创建扫码支付订单，返回支付二维码信息
- **Content-Type**：`application/json`

**请求字段**：

| 字段 | 类型 | 是否必填 | 说明 |
|------|------|---------|------|
| `name` | string | 是 | 商品名称 |
| `pay_type` | string | 否 | 支付类型，默认 `native` |
| `price` | string | 是 | 订单金额，如 `99.00` |
| `order_id` | string | 是 | 商户订单号 |
| `order_uid` | string | 否 | 用户标识 |
| `notify_url` | string | 是* | 回调通知地址，为空时取全局默认值 |
| `return_url` | string | 否 | 支付完成跳转地址 |
| `openid` | string | 否 | 用户 OpenID（JSAPI 场景） |
| `appid` | string | 否 | 小程序场景应用 ID |
| `is_mini` | bool | 否 | 是否小程序支付 |

> *`notify_url` 在请求体为空时，会自动使用 `XORPAY_NOTIFY_URL` 环境变量值。

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/pay/native \
  -H "Content-Type: application/json" \
  -d '{
    "name": "会员订阅一年期",
    "pay_type": "native",
    "price": "99.00",
    "order_id": "order-20240504-001",
    "notify_url": "https://your-domain.com/xorpay/notify"
  }'
```

**响应示例**（成功）：

```json
{
  "status": "ok",
  "aoid": "aoid_xxxxxxxxxxxx",
  "expires_in": 7200,
  "info": {"qr": "weixin://wxpay/bizpayurl?pr=xxx"}
}
```

**响应示例**（失败）：

```json
{"error": "name is required"}
```

### 6.3 创建收银台支付

- **方法**：`POST`
- **路径**：`/api/pay/cashier`
- **说明**：创建收银台支付订单，用户跳转至 XorPay 收银台完成支付
- **Content-Type**：`application/json`

**请求字段**：

| 字段 | 类型 | 是否必填 | 说明 |
|------|------|---------|------|
| `name` | string | 是 | 商品名称 |
| `pay_type` | string | 否 | 支付类型，默认 `jsapi` |
| `price` | string | 是 | 订单金额 |
| `order_id` | string | 是 | 商户订单号 |
| `order_uid` | string | 否 | 用户标识 |
| `notify_url` | string | 是* | 回调通知地址，为空时取全局默认值 |
| `return_url` | string | 否* | 支付完成跳转地址，为空时取全局默认值 |

> *`notify_url` 和 `return_url` 在请求体为空时，分别自动使用对应的环境变量值。

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/pay/cashier \
  -H "Content-Type: application/json" \
  -d '{
    "name": "会员订阅一年期",
    "pay_type": "jsapi",
    "price": "99.00",
    "order_id": "order-20240504-001",
    "return_url": "https://your-domain.com/pay/success"
  }'
```

**响应示例**（成功）：

```json
{
  "status": "ok",
  "aoid": "aoid_xxxxxxxxxxxx",
  "expires_in": 7200,
  "info": {"url": "https://xorpay.com/qr?data=xxx"}
}
```

### 6.4 创建条码支付

- **方法**：`POST`
- **路径**：`/api/pay/barcode`
- **说明**：商户扫描用户付款码，同步返回支付结果
- **Content-Type**：`application/json`

**请求字段**：

| 字段 | 类型 | 是否必填 | 说明 |
|------|------|---------|------|
| `name` | string | 是 | 商品名称 |
| `pay_type` | string | 是 | 支付类型，如 `wechat`、`alipay` |
| `price` | string | 是 | 订单金额 |
| `order_id` | string | 是 | 商户订单号 |
| `order_uid` | string | 否 | 用户标识 |
| `notify_url` | string | 是* | 回调通知地址，为空时取全局默认值 |
| `barcode` | string | 是 | 用户付款码 |

> *`notify_url` 在请求体为空时，会自动使用 `XORPAY_NOTIFY_URL` 环境变量值。

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/pay/barcode \
  -H "Content-Type: application/json" \
  -d '{
    "name": "会员订阅一年期",
    "pay_type": "wechat",
    "price": "99.00",
    "order_id": "order-20240504-001",
    "barcode": "134567890123456789"
  }'
```

**响应示例**（成功）：

```json
{
  "status": "ok",
  "aoid": "aoid_xxxxxxxxxxxx",
  "detail": {"trade_no": "xxx"}
}
```

**响应示例**（支付中）：

```json
{
  "status": "new",
  "aoid": "aoid_xxxxxxxxxxxx",
  "detail": {}
}
```

### 6.5 查询订单

- **方法**：`GET`
- **路径**：`/api/orders/:order_id`
- **说明**：根据商户订单号查询订单最新状态

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| `order_id` | string | 商户订单号 |

**请求示例**：

```bash
curl http://localhost:8080/api/orders/order-20240504-001
```

**响应示例**（已支付）：

```json
{
  "status": "success",
  "aoid": "aoid_xxxxxxxxxxxx",
  "order_id": "order-20240504-001",
  "pay_price": "99.00",
  "pay_time": "2024-05-04 10:30:00"
}
```

**响应示例**（未支付）：

```json
{
  "status": "new",
  "aoid": "aoid_xxxxxxxxxxxx",
  "order_id": "order-20240504-001"
}
```

### 6.6 发起退款

- **方法**：`POST`
- **路径**：`/api/refunds`
- **说明**：根据 XorPay 订单号发起退款
- **Content-Type**：`application/json`

**请求字段**：

| 字段 | 类型 | 是否必填 | 说明 |
|------|------|---------|------|
| `aoid` | string | 是 | XorPay 订单号 |
| `price` | string | 是 | 退款金额 |

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/refunds \
  -H "Content-Type: application/json" \
  -d '{
    "aoid": "aoid_xxxxxxxxxxxx",
    "price": "99.00"
  }'
```

**响应示例**（成功）：

```json
{
  "status": "ok",
  "aoid": "aoid_xxxxxxxxxxxx"
}
```

### 6.7 接收回调通知

- **方法**：`POST`
- **路径**：`/xorpay/notify`
- **说明**：接收 XorPay 支付结果异步通知
- **Content-Type**：`application/x-www-form-urlencoded`

**请求字段**（form 参数）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `aoid` | string | XorPay 订单号 |
| `order_id` | string | 商户订单号 |
| `pay_price` | string | 实际支付金额 |
| `pay_time` | string | 支付完成时间 |
| `sign` | string | 签名，用于验证通知合法性 |

**请求示例**：

```bash
curl -X POST http://localhost:8080/xorpay/notify \
  -d "aoid=aoid_xxxxxxxxxxxx" \
  -d "order_id=order-20240504-001" \
  -d "pay_price=99.00" \
  -d "pay_time=2024-05-04 10:30:00" \
  -d "sign=xxxxxxxxxxxxxxxx"
```

**响应**（处理成功）：

```
ok
```

**响应**（验签失败）：

```json
{"error": "bad sign"}
```

> **重要**：无论回调是否重复发送，只要订单已处理完成，都必须返回纯文本 `ok`，否则 XorPay 会持续重试。

## 7. 业务流程

### 7.1 Native 支付流程（扫码支付）

适用于 PC 网站、线下收银台等需要展示二维码的场景。

```
商户服务端                              XorPay 平台                              用户
    │                                       │                                    │
    │  1. POST /api/pay/native              │                                    │
    │ ─────────────────────────────────────>│                                    │
    │                                       │                                    │
    │  2. 返回 aoid + info.qr               │                                    │
    │ <─────────────────────────────────────│                                    │
    │                                       │                                    │
    │  3. 生成二维码展示给用户              │                                    │
    │ ────────────────────────────────────────────────────────────────────────>  │
    │                                       │                                    │
    │                                       │  4. 用户扫码并在手机端完成支付      │
    │                                       │ <───────────────────────────────── │
    │                                       │                                    │
    │  5. POST /xorpay/notify               │                                    │
    │ <─────────────────────────────────────│                                    │
    │                                       │                                    │
    │  6. 验签、更新订单状态、返回 "ok"     │                                    │
    │ ─────────────────────────────────────>│                                    │
```

**关键步骤**：

1. 调用 `/api/pay/native` 创建订单，指定 `pay_type: native`
2. 保存 `order_id` ↔ `aoid` 映射到数据库
3. 将响应中的 `info.qr` 生成二维码图片展示给用户
4. 用户使用手机扫码并在支付 App 中完成付款
5. XorPay 异步通知商户 `/xorpay/notify`
6. 商户验签通过后更新订单为"已支付"，返回 `ok`

### 7.2 Cashier 支付流程（收银台/H5）

适用于移动端 H5、网页支付等需要跳转收银台的场景。

```
商户服务端                              XorPay 平台                              用户浏览器
    │                                       │                                       │
    │  1. POST /api/pay/cashier             │                                       │
    │ ─────────────────────────────────────>│                                       │
    │                                       │                                       │
    │  2. 返回 aoid + info.url              │                                       │
    │ <─────────────────────────────────────│                                       │
    │                                       │                                       │
    │  3. 302 跳转至 XorPay 收银台          │                                       │
    │ ────────────────────────────────────────────────────────────────────────────> │
    │                                       │                                       │
    │                                       │  4. 用户在收银台完成支付              │
    │                                       │ <──────────────────────────────────── │
    │                                       │                                       │
    │  5. POST /xorpay/notify               │                                       │
    │ <─────────────────────────────────────│                                       │
    │                                       │                                       │
    │  6. 验签、更新订单状态、返回 "ok"     │                                       │
    │ ─────────────────────────────────────>│                                       │
    │                                       │                                       │
    │                                       │  7. 跳转回 return_url                 │
    │                                       │ ────────────────────────────────────> │
```

**关键步骤**：

1. 调用 `/api/pay/cashier` 创建订单，指定 `pay_type: jsapi`
2. 保存 `order_id` ↔ `aoid` 映射到数据库
3. 将用户浏览器 302 跳转至响应中的 `info.url` 或拼接的收银台地址
4. 用户在 XorPay 收银台页面选择支付方式并完成付款
5. XorPay 异步通知商户 `/xorpay/notify`
6. 商户验签通过后更新订单为"已支付"，返回 `ok`
7. 用户浏览器被重定向回创建订单时指定的 `return_url`

### 7.3 Barcode 支付流程（条码/被扫）

适用于超市、便利店等商户扫描用户付款码的场景，支付结果为同步返回。

```
商户服务端                              XorPay 平台                              用户
    │                                       │                                    │
    │  1. 用户出示付款码                    │                                    │
    │ <───────────────────────────────────────────────────────────────────────── │
    │                                       │                                    │
    │  2. POST /api/pay/barcode             │                                    │
    │ ─────────────────────────────────────>│                                    │
    │    (携带 barcode 参数)                │                                    │
    │                                       │                                    │
    │  3. 同步返回支付结果                  │                                    │
    │ <─────────────────────────────────────│                                    │
    │                                       │                                    │
    │  4. 更新订单状态                      │                                    │
    │                                       │                                    │
    │  5. POST /xorpay/notify（异步）       │                                    │
    │ <─────────────────────────────────────│                                    │
    │                                       │                                    │
    │  6. 幂等处理、返回 "ok"               │                                    │
    │ ─────────────────────────────────────>│                                    │
```

**关键步骤**：

1. 用户在支付 App 打开付款码，展示给商户
2. 商户调用 `/api/pay/barcode`，将扫描到的 `barcode` 传入
3. XorPay 同步返回支付结果（`status: ok` / `success` / `new`）
4. 商户根据同步结果更新订单状态（`new` 表示支付中，需等待回调）
5. 支付完成后 XorPay 仍会发送异步回调通知
6. 商户做幂等判断，已处理过的订单直接返回 `ok`

### 7.4 Query 查询流程

用于主动查询订单最新状态，适用于网络抖动、回调丢失等异常情况。

```
商户服务端                              XorPay 平台
    │                                       │
    │  1. GET /api/orders/:order_id         │
    │ ─────────────────────────────────────>│
    │                                       │
    │  2. 调用 XorPay Query API             │
    │ ─────────────────────────────────────>│
    │                                       │
    │  3. 返回订单最新状态                  │
    │ <─────────────────────────────────────│
    │                                       │
    │  4. 返回给调用方                      │
    │ <─────────────────────────────────────│
```

**关键步骤**：

1. 商户前端或管理后台调用 `/api/orders/:order_id`
2. 服务端使用 `client.QueryByOrderID()` 向 XorPay 发起查询
3. XorPay 返回订单实时状态（`new`、`success`、`paid` 等）
4. 服务端同步更新本地订单状态后返回给调用方

**建议**：在以下场景主动查询：
- 用户支付后长时间未收到回调
- 条码支付返回 `new`（支付中）状态后
- 管理后台需要展示实时订单状态

### 7.5 Refund 退款流程

```
商户服务端                              XorPay 平台
    │                                       │
    │  1. POST /api/refunds                 │
    │ ─────────────────────────────────────>│
    │    (携带 aoid + price)                │
    │                                       │
    │  2. 调用 XorPay Refund API            │
    │ ─────────────────────────────────────>│
    │                                       │
    │  3. 返回退款结果                      │
    │ <─────────────────────────────────────│
    │                                       │
    │  4. 更新订单状态为"已退款"            │
    │ <─────────────────────────────────────│
```

**关键步骤**：

1. 商户后台或客服系统调用 `/api/refunds`，传入 `aoid` 和退款金额 `price`
2. 服务端使用 `client.Refund()` 向 XorPay 发起退款请求
3. XorPay 返回退款结果（`status: ok`）
4. 服务端更新本地订单状态为"已退款"

**注意事项**：
- 退款金额可以小于或等于原订单金额（部分退款）
- 退款请求需要传入 XorPay 订单号 `aoid`，而非商户 `order_id`
- 建议在退款前查询订单状态，确保订单已支付成功

### 7.6 Notify 回调处理流程

XorPay 回调通知是支付闭环的关键环节，处理流程如下：

```
XorPay 平台                             商户服务端                              数据库
    │                                       │                                    │
    │  1. POST /xorpay/notify               │                                    │
    │ ─────────────────────────────────────>│                                    │
    │    (aoid, order_id, pay_price,        │                                    │
    │     pay_time, sign)                   │                                    │
    │                                       │                                    │
    │                                       │  2. 验证签名                        │
    │                                       │ ──────>                            │
    │                                       │                                    │
    │                                       │  3. 查询订单状态                    │
    │                                       │ ──────────────────────────────────>│
    │                                       │                                    │
    │                                       │  4. 返回订单状态                    │
    │                                       │ <──────────────────────────────────│
    │                                       │                                    │
    │                                       │  5. 已支付？直接返回 "ok"           │
    │                                       │  6. 未支付？更新状态为 "已支付"     │
    │                                       │ ──────────────────────────────────>│
    │                                       │                                    │
    │  7. 返回 "ok"                         │                                    │
    │ <─────────────────────────────────────│                                    │
```

**处理步骤**：

1. **接收通知**：XorPay 向 `/xorpay/notify` 发送 `application/x-www-form-urlencoded` 请求
2. **验证签名**：使用 `client.VerifyNotify(payload)` 验证签名，防止伪造通知
3. **查询订单**：根据 `order_id` 或 `aoid` 查询数据库中的订单记录
4. **幂等判断**：
   - 若订单已是"已支付"状态 → 直接返回 `ok`
   - 若订单为"待支付"状态 → 继续处理
5. **状态更新**：在数据库事务中更新订单状态为"已支付"，记录 `pay_time` 和 `pay_price`
6. **金额校验**：比对回调中的 `pay_price` 与订单原始金额是否一致
7. **返回 `ok`**：无论是否是重复通知，处理完成后必须返回纯文本 `ok`

> **重要**：如果返回非 `ok` 或 HTTP 状态码非 200，XorPay 会按照一定间隔多次重试通知。

## 8. 生产环境注意事项

### 8.1 持久化存储

示例使用内存 map 存储订单：`map[string]order`。生产环境必须替换为数据库（MySQL、PostgreSQL 等），确保：

- 订单创建后持久保存
- 服务重启不丢失数据
- 支持多实例部署共享状态

### 8.2 回调幂等性

XorPay 可能因网络等原因重复发送回调通知。生产环境必须：

1. **先验签**，再处理业务逻辑
2. 根据 `order_id` 或 `aoid` 查询订单状态
3. 若订单已是"已支付"状态，直接返回 `ok`，不再重复处理
4. 更新订单状态应在数据库事务中完成，成功后返回 `ok`

### 8.3 金额校验

收到回调后，应比对 `pay_price` 与订单原始金额是否一致，防止金额被篡改。

### 8.4 日志与监控

建议记录以下关键日志：

- 订单创建请求与响应
- 回调通知的完整参数
- 签名验证结果
- 订单状态变更

### 8.5 错误处理建议

#### 验签失败

**表现**：`client.VerifyNotify()` 返回错误，或收到签名不匹配的回调。

**处理建议**：
- 立即返回 HTTP 400，响应体不做特殊要求（XorPay 会判定为通知失败并重试）
- 记录完整请求参数和 IP 地址，排查是否为恶意伪造请求
- 检查 `XORPAY_APP_SECRET` 配置是否正确，是否与 XorPay 后台一致
- 不要基于验签失败的通知做任何业务状态变更

```go
if err := client.VerifyNotify(payload); err != nil {
    log.Printf("[SECURITY] notify sign verify failed from %s: %v", r.RemoteAddr, err)
    http.Error(w, "bad sign", http.StatusBadRequest)
    return
}
```

#### 重复通知

**表现**：同一笔订单收到多次 `/xorpay/notify` 回调，参数完全相同。

**处理建议**：
- 数据库中订单表需有明确的状态字段（如 `pending` / `paid` / `refunded`）
- 收到回调后先查询订单状态，已是 `paid` 则直接返回 `ok`
- 使用数据库唯一索引或分布式锁防止并发处理同一笔通知
- 记录通知日志时以 `aoid` + `pay_time` 做去重判断

```go
order := db.GetOrder(payload.OrderID)
if order.Status == "paid" {
    log.Printf("[IDEMPOTENT] order %s already paid, skip", payload.OrderID)
    w.Write([]byte("ok"))
    return
}
```

#### 订单不存在

**表现**：回调中的 `order_id` 在本地数据库中找不到记录。

**处理建议**：
- 返回 HTTP 404 或 400，XorPay 会重试通知
- 记录 `aoid` 和 `order_id`，人工排查是否因订单创建失败导致数据丢失
- 极端情况下可调用 `QueryByAOID` 向 XorPay 反查订单详情，补录本地数据后再处理
- 不要直接返回 `ok`，否则可能丢失这笔订单

```go
order := db.GetOrder(payload.OrderID)
if order == nil {
    log.Printf("[WARN] order not found: %s, aoid=%s", payload.OrderID, payload.AOID)
    http.Error(w, "order not found", http.StatusNotFound)
    return
}
```

#### 退款失败

**表现**：调用 `/api/refunds` 或 `client.Refund()` 返回错误，或 `status != ok`。

**处理建议**：
- 退款前必须先查询订单状态，确认订单已支付成功且未退款
- 对 `APIError` 进行细分处理：
  - `sign_error`：检查 `app_secret` 配置
  - 余额不足类错误：记录并通知运营人员
  - 订单已退款：幂等处理，视为成功
- 退款操作应记录退款流水号，便于后续对账
- 不要在前端直接暴露 XorPay 原始错误信息给普通用户

```go
resp, err := client.Refund(ctx, aoid, price)
if err != nil {
    var apiErr *xorpay.APIError
    if errors.As(err, &apiErr) {
        log.Printf("[REFUND FAILED] aoid=%s status=%s message=%s", aoid, apiErr.Status, apiErr.Message)
        // 根据 status 做差异化处理
    }
    return err
}
```

## 9. 生产环境改造清单

将 `examples/gin_app` 部署到生产环境前，建议逐条对照以下清单完成改造：

| 序号 | 改造项 | 改造内容 | 优先级 |
|------|--------|---------|--------|
| 1 | **数据库持久化** | 将内存 `map[string]order` 替换为 MySQL / PostgreSQL / Redis 等持久化存储；订单表至少包含 `order_id`、`aoid`、`name`、`price`、`status`、`created_at`、`updated_at` 字段 | P0 |
| 2 | **订单幂等** | 订单表对 `order_id` 建立唯一索引；回调处理时先查询状态，已支付直接返回 `ok`；关键更新使用数据库事务 | P0 |
| 3 | **金额校验** | 收到回调后比对 `pay_price` 与订单原始金额，不一致时记录告警并拒绝更新 | P0 |
| 4 | **签名验证** | 所有回调必须先执行 `client.VerifyNotify()` 验签，验签失败立即返回 400，不做任何业务处理 | P0 |
| 5 | **日志记录** | 记录订单创建、回调通知、退款请求、异常事件的完整上下文；建议接入 ELK / Loki 等日志系统 | P1 |
| 6 | **超时配置** | 为 `http.Client` 配置合理的超时时间（默认 10s，生产环境建议连接超时 5s、读取超时 30s） | P1 |
| 7 | **HTTPS** | 生产环境必须使用 HTTPS 部署，配置 TLS 证书；禁止明文 HTTP 传输 | P0 |
| 8 | **公网回调** | `XORPAY_NOTIFY_URL` 必须是公网可访问的 HTTPS 地址；本地开发使用 ngrok 等内网穿透工具 | P0 |
| 9 | **监控告警** | 对支付成功率、退款失败率、回调响应时间等指标配置监控和告警 | P1 |
| 10 | **限流熔断** | 对 `/api/pay/*` 接口做限流，防止刷单；对 XorPay API 调用配置熔断降级策略 | P2 |

> **P0**：必须完成，否则存在资金风险或数据丢失风险。  
> **P1**：强烈建议，提升系统可观测性和稳定性。  
> **P2**：根据业务规模选做。

## 10. 常见问题

**Q: 回调地址需要满足什么条件？**

A: 必须是外网可访问的 HTTPS 地址（XorPay 要求），且能正确处理 POST 请求并返回 `ok`。

**Q: 可以同时使用多个 `notify_url` 吗？**

A: 可以在 `NewClient` 配置中设置全局默认值，也可以在每次创建订单的请求中单独指定，请求级别优先级更高。

**Q: 条码支付为什么不需要 `return_url`？**

A: 条码支付是同步完成的，扫描用户付款码后立即返回支付结果，无需跳转页面，因此不需要 `return_url`。
