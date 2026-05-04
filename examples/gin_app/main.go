package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/warjiang/xorpay-sdk"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if isDev() {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Caller().Logger()
	} else {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
}

func isDev() bool {
	return strings.ToLower(os.Getenv("GIN_MODE")) != "release"
}

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
	_ = godotenv.Load()
	cfg := loadConfig()
	client, err := newXorPayClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create xorpay client")
	}

	router := newServer(cfg, client).router()
	log.Info().Str("addr", cfg.Addr).Msg("server starting")
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatal().Err(err).Msg("server exited")
	}
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

	var options []xorpay.Option
	if strings.TrimSpace(cfg.BaseURL) != "" {
		options = append(options, xorpay.WithBaseURL(cfg.BaseURL))
	}

	log.Info().Str("app_id", cfg.AppID).Str("base_url", cfg.BaseURL).Msg("xorpay client initialized")

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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	_ = router.SetTrustedProxies(nil)

	// recovery + structured access log
	router.Use(gin.Recovery())
	router.Use(requestIDMiddleware())
	router.Use(accessLogMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/api/pay/native", s.createNativePay)
	router.POST("/api/pay/cashier", s.createCashierPay)
	router.POST("/api/pay/barcode", s.createBarcodePay)
	router.GET("/api/orders", s.listOrders)
	router.GET("/api/orders/:order_id", s.queryOrder)
	router.POST("/api/refunds", s.refund)
	router.POST("/xorpay/notify", s.notify)

	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	return router
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

func accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		logger := log.Ctx(c.Request.Context()).With().
			Str("request_id", c.GetString("request_id")).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Logger()

		if len(c.Errors) > 0 {
			logger.Error().Strs("errors", c.Errors.Errors()).Msg("request handled with errors")
		} else if c.Writer.Status() >= 500 {
			logger.Error().Msg("request handled")
		} else if c.Writer.Status() >= 400 {
			logger.Warn().Msg("request handled")
		} else {
			logger.Info().Msg("request handled")
		}
	}
}

func ctxLog(c *gin.Context) *zerolog.Logger {
	l := log.With().Str("request_id", c.GetString("request_id")).Logger()
	return &l
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

	ctxLog(c).Info().
		Str("order_id", req.OrderID).
		Str("name", req.Name).
		Str("price", req.Price).
		Str("pay_type", req.PayType).
		Str("notify_url", req.NotifyURL).
		Msg("create native pay request")

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
		ctxLog(c).Error().Err(err).Str("order_id", req.OrderID).Msg("create native pay failed")
		xorPayError(c, err)
		return
	}

	ctxLog(c).Info().
		Str("order_id", req.OrderID).
		Str("aoid", resp.AOID).
		Str("status", resp.Status).
		Msg("create native pay success")

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

	ctxLog(c).Info().
		Str("order_id", req.OrderID).
		Str("name", req.Name).
		Str("price", req.Price).
		Str("pay_type", req.PayType).
		Msg("create cashier pay request")

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
		ctxLog(c).Error().Err(err).Str("order_id", req.OrderID).Msg("create cashier pay failed")
		xorPayError(c, err)
		return
	}

	ctxLog(c).Info().
		Str("order_id", req.OrderID).
		Str("aoid", resp.AOID).
		Str("status", resp.Status).
		Msg("create cashier pay success")

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

	ctxLog(c).Info().
		Str("order_id", req.OrderID).
		Str("name", req.Name).
		Str("price", req.Price).
		Str("pay_type", req.PayType).
		Msg("create barcode pay request")

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
		ctxLog(c).Error().Err(err).Str("order_id", req.OrderID).Msg("create barcode pay failed")
		xorPayError(c, err)
		return
	}

	ctxLog(c).Info().
		Str("order_id", req.OrderID).
		Str("aoid", resp.AOID).
		Str("status", resp.Status).
		Msg("create barcode pay success")

	s.orders.save(order{OrderID: req.OrderID, AOID: resp.AOID, Name: req.Name, PayType: req.PayType, Price: req.Price, Status: "created"})
	c.JSON(http.StatusOK, gin.H{"status": resp.Status, "aoid": resp.AOID, "detail": resp.Detail})
}

func (s *appServer) queryOrder(c *gin.Context) {
	orderID := strings.TrimSpace(c.Param("order_id"))
	if orderID == "" {
		badRequest(c, "order_id is required")
		return
	}

	ctxLog(c).Info().Str("order_id", orderID).Msg("query order")

	resp, err := s.client.QueryByOrderID(c.Request.Context(), orderID)
	if err != nil {
		ctxLog(c).Error().Err(err).Str("order_id", orderID).Msg("query order failed")
		xorPayError(c, err)
		return
	}

	if existing, ok := s.orders.get(orderID); ok && resp.AOID != "" {
		existing.AOID = resp.AOID
		existing.Status = resp.Status
		s.orders.save(existing)
	}

	ctxLog(c).Info().
		Str("order_id", orderID).
		Str("aoid", resp.AOID).
		Str("status", resp.Status).
		Msg("query order success")

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

	ctxLog(c).Info().Str("aoid", req.AOID).Str("price", req.Price).Msg("refund request")

	resp, err := s.client.Refund(c.Request.Context(), req.AOID, req.Price)
	if err != nil {
		ctxLog(c).Error().Err(err).Str("aoid", req.AOID).Msg("refund failed")
		xorPayError(c, err)
		return
	}

	ctxLog(c).Info().Str("aoid", req.AOID).Str("status", resp.Status).Msg("refund success")
	c.JSON(http.StatusOK, resp)
}

func (s *appServer) notify(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		badRequest(c, err.Error())
		return
	}

	ctxLog(c).Info().Str("form", c.Request.Form.Encode()).Msg("notify received")

	payload := xorpay.NotifyPayloadFromValues(c.Request.Form)
	if err := s.client.VerifyNotify(payload); err != nil {
		ctxLog(c).Error().Err(err).Msg("notify verify failed")
		badRequest(c, "bad sign")
		return
	}

	if _, err := s.orders.markPaid(payload); err != nil {
		ctxLog(c).Error().Err(err).Str("order_id", payload.OrderID).Msg("notify mark paid failed")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctxLog(c).Info().
		Str("order_id", payload.OrderID).
		Str("aoid", payload.AOID).
		Str("pay_price", payload.PayPrice).
		Str("pay_time", payload.PayTime).
		Msg("notify handled")

	c.String(http.StatusOK, "ok")
}

func bindJSON(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		badRequest(c, err.Error())
		return false
	}
	return true
}

func validateBase(name, payType, price, orderID, notifyURL string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(payType) == "" {
		return errors.New("pay_type is required")
	}
	if strings.TrimSpace(price) == "" {
		return errors.New("price is required")
	}
	if strings.TrimSpace(orderID) == "" {
		return errors.New("order_id is required")
	}
	if strings.TrimSpace(notifyURL) == "" {
		return errors.New("notify_url or XORPAY_NOTIFY_URL is required")
	}
	return nil
}

func validatePayRequest(req payRequest) error {
	return validateBase(req.Name, req.PayType, req.Price, req.OrderID, req.NotifyURL)
}

func validateBarcodeRequest(req barcodeRequest) error {
	if err := validateBase(req.Name, req.PayType, req.Price, req.OrderID, req.NotifyURL); err != nil {
		return err
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

func (s *orderStore) all() []order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]order, 0, len(s.orders))
	for _, o := range s.orders {
		result = append(result, o)
	}
	return result
}

func (s *appServer) listOrders(c *gin.Context) {
	c.JSON(http.StatusOK, s.orders.all())
}
