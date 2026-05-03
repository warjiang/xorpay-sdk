package xorpay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Option func(*Client)

type Client struct {
	appID     string
	appSecret string
	baseURL   string
	http      Doer
	userAgent string
}

func WithHTTPClient(client Doer) Option {
	return func(c *Client) {
		if client != nil {
			c.http = client
		}
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if strings.TrimSpace(baseURL) != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		if strings.TrimSpace(userAgent) != "" {
			c.userAgent = userAgent
		}
	}
}

func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if strings.TrimSpace(cfg.AppID) == "" {
		return nil, errors.New("appid is required")
	}
	if strings.TrimSpace(cfg.AppSecret) == "" {
		return nil, errors.New("app secret is required")
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	c := &Client{
		appID:     cfg.AppID,
		appSecret: cfg.AppSecret,
		baseURL:   strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "xorpay-go-sdk/0.1",
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.http == nil {
		return nil, errors.New("http client is nil")
	}
	return c, nil
}

func (c *Client) CreatePay(ctx context.Context, req PayRequest) (*PayResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	form := req.toValues()
	form.Set("sign", SignPay(req.Name, req.PayType, req.Price, req.OrderID, req.NotifyURL, c.appSecret))

	var resp PayResponse
	data, err := c.doJSON(ctx, http.MethodPost, "/api/pay/"+c.appID, form, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, &APIError{Endpoint: "/api/pay", Status: resp.Status, Message: parseAPIMessage(data)}
	}
	return &resp, nil
}

func (c *Client) CreateCashier(ctx context.Context, req CashierRequest) (*CashierResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	form := req.toValues()
	form.Set("sign", SignPay(req.Name, req.PayType, req.Price, req.OrderID, req.NotifyURL, c.appSecret))

	var resp CashierResponse
	data, err := c.doJSON(ctx, http.MethodPost, "/api/cashier/"+c.appID, form, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, &APIError{Endpoint: "/api/cashier", Status: resp.Status, Message: parseAPIMessage(data)}
	}
	return &resp, nil
}

func (c *Client) CreateBarcodePay(ctx context.Context, req BarcodePayRequest) (*BarcodePayResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	form := req.toValues()
	form.Set("sign", SignBarcodePay(req.Name, req.PayType, req.Price, req.OrderID, req.NotifyURL, req.Barcode, c.appSecret))

	var resp BarcodePayResponse
	data, err := c.doJSON(ctx, http.MethodPost, "/api/barcode_pay/"+c.appID, form, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" && resp.Status != "success" && resp.Status != "new" {
		return nil, &APIError{Endpoint: "/api/barcode_pay", Status: resp.Status, Message: parseAPIMessage(data)}
	}
	return &resp, nil
}

func (c *Client) QueryByAOID(ctx context.Context, aoid string) (*QueryResponse, error) {
	aoid = strings.TrimSpace(aoid)
	if aoid == "" {
		return nil, errors.New("aoid is required")
	}

	var resp QueryResponse
	if _, err := c.doJSON(ctx, http.MethodGet, "/api/query/"+aoid, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) QueryByOrderID(ctx context.Context, orderID string) (*QueryResponse, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, errors.New("order_id is required")
	}

	params := url.Values{}
	params.Set("order_id", orderID)
	params.Set("sign", SignQueryByOrderID(orderID, c.appSecret))

	var resp QueryResponse
	if _, err := c.doJSON(ctx, http.MethodGet, "/api/query2/"+c.appID, params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Refund(ctx context.Context, aoid, price string) (*RefundResponse, error) {
	aoid = strings.TrimSpace(aoid)
	price = strings.TrimSpace(price)
	if aoid == "" {
		return nil, errors.New("aoid is required")
	}
	if price == "" {
		return nil, errors.New("price is required")
	}

	params := url.Values{}
	params.Set("price", price)
	params.Set("sign", SignRefund(price, c.appSecret))

	var resp RefundResponse
	data, err := c.doJSON(ctx, http.MethodPost, "/api/refund/"+aoid, params, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, &APIError{Endpoint: "/api/refund", Status: resp.Status, Message: parseAPIMessage(data)}
	}
	return &resp, nil
}

func (c *Client) BuildOpenIDURL(callback string) (string, error) {
	callback = strings.TrimSpace(callback)
	if callback == "" {
		return "", errors.New("callback is required")
	}

	u, err := url.Parse(c.baseURL + "/api/openid/" + c.appID)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("callback", callback)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) BuildQRURL(data string) string {
	u, _ := url.Parse(c.baseURL + "/qr")
	q := u.Query()
	q.Set("data", data)
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) VerifyNotify(payload NotifyPayload) error {
	if err := payload.Validate(); err != nil {
		return err
	}
	expected := SignNotify(payload.AOID, payload.OrderID, payload.PayPrice, payload.PayTime, c.appSecret)
	if !strings.EqualFold(expected, payload.Sign) {
		return &APIError{Endpoint: "notify", Status: "sign_error", Message: "notify sign mismatch"}
	}
	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, params url.Values, out any) ([]byte, error) {
	fullURL := c.baseURL + path
	var body io.Reader
	if method == http.MethodGet && len(params) > 0 {
		fullURL += "?" + params.Encode()
	} else if method == http.MethodPost {
		body = bytes.NewBufferString(params.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{Endpoint: path, HTTPStatus: resp.StatusCode, RawBody: string(data), Message: "http status >= 400"}
	}

	if err := json.Unmarshal(data, out); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return data, nil
}

func parseAPIMessage(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	if info, ok := payload["info"]; ok {
		if text, ok := info.(string); ok {
			return text
		}
	}
	if detail, ok := payload["detail"]; ok {
		if text, ok := detail.(string); ok {
			return text
		}
	}
	return ""
}
