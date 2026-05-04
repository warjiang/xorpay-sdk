package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	xorpay "github.com/warjiang/xorpay-sdk"
)

type appConfig struct {
	AppID     string
	Secret    string
	NotifyURL string
	ReturnURL string
	BaseURL   string
	Addr      string
}

type appServer struct {
	cfg    appConfig
	client *xorpay.Client
	orders *orderStore
}

type order struct {
	OrderID string `json:"order_id"`
	AOID    string `json:"aoid,omitempty"`
	Name    string `json:"name"`
	PayType string `json:"pay_type"`
	Price   string `json:"price"`
	Status  string `json:"status"`
}

type orderStore struct {
	mu     sync.RWMutex
	orders map[string]order
}

type payRequest struct {
	Name      string `json:"name"`
	PayType   string `json:"pay_type"`
	Price     string `json:"price"`
	OrderID   string `json:"order_id"`
	OrderUID  string `json:"order_uid"`
	NotifyURL string `json:"notify_url"`
	ReturnURL string `json:"return_url"`
	OpenID    string `json:"openid"`
	AppID     string `json:"appid"`
	IsMini    bool   `json:"is_mini"`
}

type barcodeRequest struct {
	Name      string `json:"name"`
	PayType   string `json:"pay_type"`
	Price     string `json:"price"`
	OrderID   string `json:"order_id"`
	OrderUID  string `json:"order_uid"`
	NotifyURL string `json:"notify_url"`
	Barcode   string `json:"barcode"`
}

type refundRequest struct {
	AOID  string `json:"aoid"`
	Price string `json:"price"`
}

func main() {
	cfg := loadConfig()
	client, err := newXorPayClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := newServer(cfg, client).router()
	log.Printf("listen on %s", cfg.Addr)
	log.Fatal(router.Run(cfg.Addr))
}

func loadConfig() appConfig {
	return appConfig{
		AppID:     os.Getenv("XORPAY_APP_ID"),
		Secret:    os.Getenv("XORPAY_APP_SECRET"),
		NotifyURL: os.Getenv("XORPAY_NOTIFY_URL"),
		ReturnURL: os.Getenv("XORPAY_RETURN_URL"),
		BaseURL:   os.Getenv("XORPAY_BASE_URL"),
		Addr:      envOr("GIN_ADDR", ":8080"),
	}
}

func envOr(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func newXorPayClient(cfg appConfig) (*xorpay.Client, error) {
	if strings.TrimSpace(cfg.AppID) == "" {
		return nil, errors.New("XORPAY_APP_ID is required")
	}
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, errors.New("XORPAY_APP_SECRET is required")
	}

	options := []xorpay.Option{}
	if strings.TrimSpace(cfg.BaseURL) != "" {
		options = append(options, xorpay.WithBaseURL(cfg.BaseURL))
	}

	return xorpay.NewClient(xorpay.Config{
		AppID:     cfg.AppID,
		AppSecret: cfg.Secret,
	}, options...)
}

func newServer(cfg appConfig, client *xorpay.Client) *appServer {
	return &appServer{
		cfg:    cfg,
		client: client,
		orders: &orderStore{orders: map[string]order{}},
	}
}

func (s *appServer) router() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/api/pay/native", s.createNativePay)
	router.POST("/api/pay/cashier", s.createCashierPay)
	router.POST("/api/pay/barcode", s.createBarcodePay)
	router.GET("/api/orders/:order_id", s.queryOrder)
	router.POST("/api/refunds", s.refund)
	router.POST("/xorpay/notify", s.notify)

	return router
}

func (s *appServer) createNativePay(c *gin.Context) {
	var req payRequest
	if !bindJSON(c, &req) {
		return
	}
	req.PayType = defaultString(req.PayType, "native")
	req.NotifyURL = defaultString(req.NotifyURL, s.cfg.NotifyURL)

	if err := validatePayRequest(req); err != nil {
		badRequest(c, err.Error())
		return
	}

	resp, err := s.client.CreatePay(c.Request.Context(), xorpay.PayRequest{
		Name:      req.Name,
		PayType:   req.PayType,
		Price:     req.Price,
		OrderID:   req.OrderID,
		OrderUID:  req.OrderUID,
		NotifyURL: req.NotifyURL,
		ReturnURL: req.ReturnURL,
		OpenID:    req.OpenID,
		AppID:     req.AppID,
		IsMini:    req.IsMini,
	})
	if err != nil {
		xorPayError(c, err)
		return
	}

	s.orders.save(orderFromPay(req, resp.AOID, "created"))
	c.JSON(http.StatusOK, gin.H{"status": resp.Status, "aoid": resp.AOID, "expires_in": resp.ExpiresIn, "info": resp.Info})
}

func (s *appServer) createCashierPay(c *gin.Context) {
	var req payRequest
	if !bindJSON(c, &req) {
		return
	}
	req.PayType = defaultString(req.PayType, "jsapi")
	req.NotifyURL = defaultString(req.NotifyURL, s.cfg.NotifyURL)
	req.ReturnURL = defaultString(req.ReturnURL, s.cfg.ReturnURL)

	if err := validatePayRequest(req); err != nil {
		badRequest(c, err.Error())
		return
	}

	resp, err := s.client.CreateCashier(c.Request.Context(), xorpay.CashierRequest{
		Name:      req.Name,
		PayType:   req.PayType,
		Price:     req.Price,
		OrderID:   req.OrderID,
		OrderUID:  req.OrderUID,
		NotifyURL: req.NotifyURL,
		ReturnURL: req.ReturnURL,
	})
	if err != nil {
		xorPayError(c, err)
		return
	}

	s.orders.save(orderFromPay(req, resp.AOID, "created"))
	c.JSON(http.StatusOK, gin.H{"status": resp.Status, "aoid": resp.AOID, "expires_in": resp.ExpiresIn, "info": resp.Info})
}

func (s *appServer) createBarcodePay(c *gin.Context) {
	var req barcodeRequest
	if !bindJSON(c, &req) {
		return
	}
	req.NotifyURL = defaultString(req.NotifyURL, s.cfg.NotifyURL)

	if err := validateBarcodeRequest(req); err != nil {
		badRequest(c, err.Error())
		return
	}

	resp, err := s.client.CreateBarcodePay(c.Request.Context(), xorpay.BarcodePayRequest{
		Name:      req.Name,
		PayType:   req.PayType,
		Price:     req.Price,
		OrderID:   req.OrderID,
		OrderUID:  req.OrderUID,
		NotifyURL: req.NotifyURL,
		Barcode:   req.Barcode,
	})
	if err != nil {
		xorPayError(c, err)
		return
	}

	s.orders.save(order{OrderID: req.OrderID, AOID: resp.AOID, Name: req.Name, PayType: req.PayType, Price: req.Price, Status: "created"})
	c.JSON(http.StatusOK, gin.H{"status": resp.Status, "aoid": resp.AOID, "detail": resp.Detail})
}

func (s *appServer) queryOrder(c *gin.Context) {
	orderID := strings.TrimSpace(c.Param("order_id"))
	if orderID == "" {
		badRequest(c, "order_id is required")
		return
	}

	resp, err := s.client.QueryByOrderID(c.Request.Context(), orderID)
	if err != nil {
		xorPayError(c, err)
		return
	}

	if existing, ok := s.orders.get(orderID); ok && resp.AOID != "" {
		existing.AOID = resp.AOID
		existing.Status = resp.Status
		s.orders.save(existing)
	}

	c.JSON(http.StatusOK, resp)
}

func (s *appServer) refund(c *gin.Context) {
	var req refundRequest
	if !bindJSON(c, &req) {
		return
	}
	if strings.TrimSpace(req.AOID) == "" || strings.TrimSpace(req.Price) == "" {
		badRequest(c, "aoid and price are required")
		return
	}

	resp, err := s.client.Refund(c.Request.Context(), req.AOID, req.Price)
	if err != nil {
		xorPayError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (s *appServer) notify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		badRequest(c, err.Error())
		return
	}

	payload := xorpay.NotifyPayloadFromValues(c.Request.Form)
	if err := s.client.VerifyNotify(payload); err != nil {
		badRequest(c, "bad sign")
		return
	}

	if _, err := s.orders.markPaid(payload); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	log.Printf("xorpay notify order_id=%s aoid=%s pay_price=%s pay_time=%s", payload.OrderID, payload.AOID, payload.PayPrice, payload.PayTime)
	c.String(http.StatusOK, "ok")
}

func bindJSON(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		badRequest(c, err.Error())
		return false
	}
	return true
}

func validatePayRequest(req payRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.PayType) == "" {
		return errors.New("pay_type is required")
	}
	if strings.TrimSpace(req.Price) == "" {
		return errors.New("price is required")
	}
	if strings.TrimSpace(req.OrderID) == "" {
		return errors.New("order_id is required")
	}
	if strings.TrimSpace(req.NotifyURL) == "" {
		return errors.New("notify_url or XORPAY_NOTIFY_URL is required")
	}
	return nil
}

func validateBarcodeRequest(req barcodeRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.PayType) == "" {
		return errors.New("pay_type is required")
	}
	if strings.TrimSpace(req.Price) == "" {
		return errors.New("price is required")
	}
	if strings.TrimSpace(req.OrderID) == "" {
		return errors.New("order_id is required")
	}
	if strings.TrimSpace(req.NotifyURL) == "" {
		return errors.New("notify_url or XORPAY_NOTIFY_URL is required")
	}
	if strings.TrimSpace(req.Barcode) == "" {
		return errors.New("barcode is required")
	}
	return nil
}

func orderFromPay(req payRequest, aoid, status string) order {
	return order{OrderID: req.OrderID, AOID: aoid, Name: req.Name, PayType: req.PayType, Price: req.Price, Status: status}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func xorPayError(c *gin.Context, err error) {
	var apiErr *xorpay.APIError
	if errors.As(err, &apiErr) {
		c.JSON(http.StatusBadGateway, gin.H{"error": apiErr.Error(), "status": apiErr.Status})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func (s *orderStore) save(value order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[value.OrderID] = value
}

func (s *orderStore) get(orderID string) (order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.orders[orderID]
	return value, ok
}

func (s *orderStore) markPaid(payload xorpay.NotifyPayload) (order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.orders[payload.OrderID]
	if !ok {
		return order{}, fmt.Errorf("order %s not found", payload.OrderID)
	}
	if value.Price != payload.PayPrice {
		return order{}, fmt.Errorf("pay_price mismatch for order %s", payload.OrderID)
	}

	value.AOID = payload.AOID
	value.Status = "paid"
	s.orders[payload.OrderID] = value
	return value, nil
}
