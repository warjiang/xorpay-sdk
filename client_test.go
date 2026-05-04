package xorpay

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	c, err := NewClient(Config{AppID: "704046", AppSecret: "8d15136f11f3458a91dfe84a0145c612"}, WithBaseURL(ts.URL))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	return c, ts
}

func TestCreatePay(t *testing.T) {
	client, ts := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/api/pay/704046" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm error: %v", err)
		}

		expectedSign := SignPay(
			r.Form.Get("name"),
			r.Form.Get("pay_type"),
			r.Form.Get("price"),
			r.Form.Get("order_id"),
			r.Form.Get("notify_url"),
			"8d15136f11f3458a91dfe84a0145c612",
		)
		if got := r.Form.Get("sign"); got != expectedSign {
			t.Fatalf("sign mismatch: got=%s want=%s", got, expectedSign)
		}

		_, _ = w.Write([]byte(`{"status":"ok","aoid":"aoid123","expires_in":7200,"info":{"qr":"weixin://abc"}}`))
	})
	defer ts.Close()

	resp, err := client.CreatePay(context.Background(), PayRequest{
		Name:      "内容订阅一年期",
		PayType:   "native",
		Price:     "50.00",
		OrderID:   "demo-3",
		NotifyURL: "http://abc.com/xorpay_notify",
	})
	if err != nil {
		t.Fatalf("CreatePay error: %v", err)
	}
	if resp.AOID != "aoid123" || resp.Status != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreatePayAPIError(t *testing.T) {
	client, ts := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"sign_error"}`))
	})
	defer ts.Close()

	_, err := client.CreatePay(context.Background(), PayRequest{
		Name:      "A",
		PayType:   "native",
		Price:     "1.00",
		OrderID:   "o1",
		NotifyURL: "http://notify",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsStatus(err, "sign_error") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQueryByOrderID(t *testing.T) {
	client, ts := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/api/query2/704046" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("order_id") != "demo-order" {
			t.Fatalf("order_id=%s", q.Get("order_id"))
		}
		expectedSign := SignQueryByOrderID("demo-order", "8d15136f11f3458a91dfe84a0145c612")
		if q.Get("sign") != expectedSign {
			t.Fatalf("sign mismatch: got=%s want=%s", q.Get("sign"), expectedSign)
		}
		_, _ = w.Write([]byte(`{"status":"success","aoid":"a1","order_id":"demo-order"}`))
	})
	defer ts.Close()

	resp, err := client.QueryByOrderID(context.Background(), "demo-order")
	if err != nil {
		t.Fatalf("QueryByOrderID error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("status=%s", resp.Status)
	}
}

func TestRefund(t *testing.T) {
	client, ts := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/refund/aoid-1" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm error: %v", err)
		}
		if r.Form.Get("price") != "0.01" {
			t.Fatalf("price=%s", r.Form.Get("price"))
		}
		expected := SignRefund("0.01", "8d15136f11f3458a91dfe84a0145c612")
		if r.Form.Get("sign") != expected {
			t.Fatalf("sign mismatch")
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	defer ts.Close()

	if _, err := client.Refund(context.Background(), "aoid-1", "0.01"); err != nil {
		t.Fatalf("Refund error: %v", err)
	}
}

func TestBuildOpenIDAndQR(t *testing.T) {
	client, err := NewClient(Config{AppID: "704046", AppSecret: "secret"}, WithBaseURL("https://xorpay.com"))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	openidURL, err := client.BuildOpenIDURL("https://merchant.test/callback?a=1")
	if err != nil {
		t.Fatalf("BuildOpenIDURL error: %v", err)
	}
	u, err := url.Parse(openidURL)
	if err != nil {
		t.Fatalf("url parse error: %v", err)
	}
	if u.Path != "/api/openid/704046" {
		t.Fatalf("path=%s", u.Path)
	}
	if u.Query().Get("callback") != "https://merchant.test/callback?a=1" {
		t.Fatalf("callback=%s", u.Query().Get("callback"))
	}

	qrURL := client.BuildQRURL("weixin://wxpay/bizpayurl?pr=abc")
	if !strings.Contains(qrURL, "/qr?") {
		t.Fatalf("qrURL=%s", qrURL)
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("XORPAY_APP_ID", "env-id")
	t.Setenv("XORPAY_APP_SECRET", "env-secret")
	t.Setenv("XORPAY_NOTIFY_URL", "https://example.com/notify")
	t.Setenv("XORPAY_RETURN_URL", "https://example.com/return")

	cfg := ConfigFromEnv()
	if cfg.AppID != "env-id" {
		t.Fatalf("appid=%s", cfg.AppID)
	}
	if cfg.AppSecret != "env-secret" {
		t.Fatalf("secret=%s", cfg.AppSecret)
	}
	if cfg.NotifyURL != "https://example.com/notify" {
		t.Fatalf("notify_url=%s", cfg.NotifyURL)
	}
	if cfg.ReturnURL != "https://example.com/return" {
		t.Fatalf("return_url=%s", cfg.ReturnURL)
	}
}

func TestDoJSONBadResponse(t *testing.T) {
	client, ts := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("gateway error"))
	})
	defer ts.Close()

	_, err := client.QueryByAOID(context.Background(), "a1")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.HTTPStatus != http.StatusBadGateway {
		t.Fatalf("http status=%d", apiErr.HTTPStatus)
	}
}
