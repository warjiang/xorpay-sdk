package xorpay

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const DefaultBaseURL = "https://xorpay.com"

type Config struct {
	AppID     string
	AppSecret string
	BaseURL   string
	NotifyURL string
	ReturnURL string
}

type PayRequest struct {
	Name      string
	PayType   string
	Price     string
	OrderID   string
	OrderUID  string
	NotifyURL string
	ReturnURL string
	OpenID    string
	AppID     string
	IsMini    bool
}

func (r PayRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(r.PayType) == "" {
		return fmt.Errorf("pay_type is required")
	}
	if strings.TrimSpace(r.Price) == "" {
		return fmt.Errorf("price is required")
	}
	if strings.TrimSpace(r.OrderID) == "" {
		return fmt.Errorf("order_id is required")
	}
	if strings.TrimSpace(r.NotifyURL) == "" {
		return fmt.Errorf("notify_url is required")
	}
	return nil
}

func (r PayRequest) toValues() url.Values {
	v := url.Values{}
	v.Set("name", r.Name)
	v.Set("pay_type", r.PayType)
	v.Set("price", r.Price)
	v.Set("order_id", r.OrderID)
	if r.OrderUID != "" {
		v.Set("order_uid", r.OrderUID)
	}
	v.Set("notify_url", r.NotifyURL)
	if r.ReturnURL != "" {
		v.Set("return_url", r.ReturnURL)
	}
	if r.OpenID != "" {
		v.Set("openid", r.OpenID)
	}
	if r.AppID != "" {
		v.Set("appid", r.AppID)
	}
	if r.IsMini {
		v.Set("is_mini", "true")
	}
	return v
}

type CashierRequest struct {
	Name      string
	PayType   string
	Price     string
	OrderID   string
	OrderUID  string
	NotifyURL string
	ReturnURL string
}

func (r CashierRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(r.PayType) == "" {
		return fmt.Errorf("pay_type is required")
	}
	if strings.TrimSpace(r.Price) == "" {
		return fmt.Errorf("price is required")
	}
	if strings.TrimSpace(r.OrderID) == "" {
		return fmt.Errorf("order_id is required")
	}
	if strings.TrimSpace(r.NotifyURL) == "" {
		return fmt.Errorf("notify_url is required")
	}
	return nil
}

func (r CashierRequest) toValues() url.Values {
	v := url.Values{}
	v.Set("name", r.Name)
	v.Set("pay_type", r.PayType)
	v.Set("price", r.Price)
	v.Set("order_id", r.OrderID)
	if r.OrderUID != "" {
		v.Set("order_uid", r.OrderUID)
	}
	v.Set("notify_url", r.NotifyURL)
	if r.ReturnURL != "" {
		v.Set("return_url", r.ReturnURL)
	}
	return v
}

type BarcodePayRequest struct {
	Name      string
	PayType   string
	Price     string
	OrderID   string
	OrderUID  string
	NotifyURL string
	Barcode   string
}

func (r BarcodePayRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(r.PayType) == "" {
		return fmt.Errorf("pay_type is required")
	}
	if strings.TrimSpace(r.Price) == "" {
		return fmt.Errorf("price is required")
	}
	if strings.TrimSpace(r.OrderID) == "" {
		return fmt.Errorf("order_id is required")
	}
	if strings.TrimSpace(r.NotifyURL) == "" {
		return fmt.Errorf("notify_url is required")
	}
	if strings.TrimSpace(r.Barcode) == "" {
		return fmt.Errorf("barcode is required")
	}
	return nil
}

func (r BarcodePayRequest) toValues() url.Values {
	v := url.Values{}
	v.Set("name", r.Name)
	v.Set("pay_type", r.PayType)
	v.Set("price", r.Price)
	v.Set("order_id", r.OrderID)
	if r.OrderUID != "" {
		v.Set("order_uid", r.OrderUID)
	}
	v.Set("notify_url", r.NotifyURL)
	v.Set("barcode", r.Barcode)
	return v
}

type PayResponse struct {
	Status    string          `json:"status"`
	AOID      string          `json:"aoid,omitempty"`
	ExpiresIn int             `json:"expires_in,omitempty"`
	Info      json.RawMessage `json:"info,omitempty"`
	Detail    json.RawMessage `json:"detail,omitempty"`
}

type CashierResponse struct {
	Status    string          `json:"status"`
	AOID      string          `json:"aoid,omitempty"`
	ExpiresIn int             `json:"expires_in,omitempty"`
	Info      json.RawMessage `json:"info,omitempty"`
	Detail    json.RawMessage `json:"detail,omitempty"`
}

type BarcodePayResponse struct {
	Status string          `json:"status"`
	AOID   string          `json:"aoid,omitempty"`
	Detail json.RawMessage `json:"detail,omitempty"`
}

type QueryResponse struct {
	Status   string          `json:"status"`
	AOID     string          `json:"aoid,omitempty"`
	OrderID  string          `json:"order_id,omitempty"`
	PayPrice string          `json:"pay_price,omitempty"`
	PayTime  string          `json:"pay_time,omitempty"`
	Detail   json.RawMessage `json:"detail,omitempty"`
	Info     json.RawMessage `json:"info,omitempty"`
}

type RefundResponse struct {
	Status string          `json:"status"`
	AOID   string          `json:"aoid,omitempty"`
	Info   json.RawMessage `json:"info,omitempty"`
	Detail json.RawMessage `json:"detail,omitempty"`
}

type NotifyPayload struct {
	AOID     string
	OrderID  string
	PayPrice string
	PayTime  string
	Sign     string
}

func (n NotifyPayload) Validate() error {
	if strings.TrimSpace(n.AOID) == "" {
		return fmt.Errorf("aoid is required")
	}
	if strings.TrimSpace(n.OrderID) == "" {
		return fmt.Errorf("order_id is required")
	}
	if strings.TrimSpace(n.PayPrice) == "" {
		return fmt.Errorf("pay_price is required")
	}
	if strings.TrimSpace(n.PayTime) == "" {
		return fmt.Errorf("pay_time is required")
	}
	if strings.TrimSpace(n.Sign) == "" {
		return fmt.Errorf("sign is required")
	}
	return nil
}
