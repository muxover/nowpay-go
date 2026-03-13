// Package nowpay provides a Go SDK for the NOWPayments API (NowPay Go).
package nowpay

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/muxover/nowpay-go/internal"
	"github.com/muxover/nowpay-go/models"
)

// Sentinel errors for API and client failures. Use errors.Is(err, nowpay.ErrX) to check.
var (
	ErrInvalidAPIKey        = internal.ErrInvalidAPIKey
	ErrPaymentNotFound      = internal.ErrPaymentNotFound
	ErrInvoiceNotFound      = internal.ErrInvoiceNotFound
	ErrPayoutNotFound       = internal.ErrPayoutNotFound
	ErrSubscriptionNotFound = internal.ErrSubscriptionNotFound
	ErrRateLimited          = internal.ErrRateLimited
	ErrInvalidSignature     = internal.ErrInvalidSignature
	ErrBadRequest           = internal.ErrBadRequest
	ErrValidation           = internal.ErrValidation
	ErrServerError          = internal.ErrServerError
	ErrUnauthorized         = internal.ErrUnauthorized
	ErrForbidden            = internal.ErrForbidden
	ErrNotFound             = internal.ErrNotFound
)

// IsAPIError reports whether err is or wraps an API error (status code + message).
func IsAPIError(err error) bool { return internal.IsAPIError(err) }

const DefaultBaseURL = "https://api.nowpayments.io/v1"

// RetryConfig configures retry behavior for 5xx and 429 responses. Zero value means no retries.
type RetryConfig struct {
	MaxRetries     int           // max attempts (0 = no retries, 3 recommended)
	InitialBackoff time.Duration // first backoff (e.g. 1*time.Second)
	MaxBackoff     time.Duration // cap (e.g. 30*time.Second)
}

// Config configures the NowPay client.
type Config struct {
	APIKey     string
	Token      string        // optional JWT (from Auth); required for List payments / Create payout on some accounts
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
	Retry      RetryConfig   // optional retry for 5xx and 429
}

// Client is the main NowPay API client.
type Client struct {
	Payments      *PaymentsService
	Invoices      *InvoicesService
	Currencies    *CurrenciesService
	Estimates     *EstimatesService
	Payouts       *PayoutsService
	Subscriptions *SubscriptionsService
	do            *internal.Client
}

// NewClient creates a new NowPay client. APIKey is required.
func NewClient(cfg Config) *Client {
	if cfg.APIKey == "" {
		panic("nowpay: APIKey is required")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	do := internal.NewClient(baseURL, cfg.APIKey, cfg.Token, timeout, cfg.HTTPClient, internal.RetryConfig{
		MaxRetries:     cfg.Retry.MaxRetries,
		InitialBackoff: cfg.Retry.InitialBackoff,
		MaxBackoff:     cfg.Retry.MaxBackoff,
	})
	c := &Client{do: do}
	c.Payments = &PaymentsService{c: c}
	c.Invoices = &InvoicesService{c: c}
	c.Currencies = &CurrenciesService{c: c}
	c.Estimates = &EstimatesService{c: c}
	c.Payouts = &PayoutsService{c: c}
	c.Subscriptions = &SubscriptionsService{c: c}
	return c
}

// Status checks API availability (GET /v1/status).
func (c *Client) Status(ctx context.Context) (*models.StatusResponse, error) {
	var out models.StatusResponse
	err := c.do.Do(ctx, "GET", "/v1/status", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Auth obtains a JWT token (POST /v1/auth). Required for List payments and Create payout on some accounts.
func (c *Client) Auth(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	var out models.AuthResponse
	err := c.do.Do(ctx, "POST", "/v1/auth", &models.AuthRequest{Email: email, Password: password}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance returns balance per currency (GET /v1/balance).
func (c *Client) Balance(ctx context.Context) (*models.BalanceResponse, error) {
	var out models.BalanceResponse
	err := c.do.Do(ctx, "GET", "/v1/balance", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PaymentsService handles payment operations.
type PaymentsService struct{ c *Client }

// Create creates a new payment.
func (s *PaymentsService) Create(ctx context.Context, req *models.CreatePaymentRequest) (*models.Payment, error) {
	var out models.Payment
	err := s.c.do.Do(ctx, "POST", "/v1/payment", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a payment by ID.
func (s *PaymentsService) Get(ctx context.Context, paymentID int64) (*models.Payment, error) {
	var out models.Payment
	err := s.c.do.Do(ctx, "GET", "/v1/payment/"+formatInt64(paymentID), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a list of payments. Pass nil for default params.
func (s *PaymentsService) List(ctx context.Context, params *models.ListPaymentsParams) ([]models.Payment, error) {
	path := "/v1/payment"
	if params != nil {
		if q := encodeListPaymentsParams(params); q != "" {
			path += "?" + q
		}
	}
	var out []models.Payment
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetFlow returns detailed payment flow (e.g. confirmations) for a payment.
func (s *PaymentsService) GetFlow(ctx context.Context, paymentID int64) (*models.PaymentFlow, error) {
	var out models.PaymentFlow
	err := s.c.do.Do(ctx, "GET", "/v1/payment/"+formatInt64(paymentID)+"/flow", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Refund requests a refund for a payment.
func (s *PaymentsService) Refund(ctx context.Context, paymentID int64, req *models.RefundRequest) (*models.RefundResponse, error) {
	if req == nil {
		req = &models.RefundRequest{PaymentID: paymentID}
	} else if req.PaymentID == 0 {
		req.PaymentID = paymentID
	}
	var out models.RefundResponse
	err := s.c.do.Do(ctx, "POST", "/v1/payment/"+formatInt64(paymentID)+"/refund", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateFromInvoice creates a payment from an invoice (POST /v1/invoice-payment).
func (s *PaymentsService) CreateFromInvoice(ctx context.Context, req *models.CreateInvoicePaymentRequest) (*models.Payment, error) {
	var out models.Payment
	err := s.c.do.Do(ctx, "POST", "/v1/invoice-payment", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMerchantEstimate gets/updates the payment estimate (POST /v1/payment/:id/update-merchant-estimate).
func (s *PaymentsService) UpdateMerchantEstimate(ctx context.Context, paymentID int64) (*models.UpdateMerchantEstimateResponse, error) {
	var out models.UpdateMerchantEstimateResponse
	err := s.c.do.Do(ctx, "POST", "/v1/payment/"+formatInt64(paymentID)+"/update-merchant-estimate", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// InvoicesService handles invoice operations.
type InvoicesService struct{ c *Client }

// Create creates a new invoice.
func (s *InvoicesService) Create(ctx context.Context, req *models.CreateInvoiceRequest) (*models.Invoice, error) {
	var out models.Invoice
	err := s.c.do.Do(ctx, "POST", "/v1/invoice", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns an invoice by ID.
func (s *InvoicesService) Get(ctx context.Context, invoiceID string) (*models.Invoice, error) {
	var out models.Invoice
	err := s.c.do.Do(ctx, "GET", "/v1/invoice/"+invoiceID, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrenciesService handles currency operations.
type CurrenciesService struct{ c *Client }

// Supported returns the list of supported currencies (GET /v1/currencies).
func (s *CurrenciesService) Supported(ctx context.Context) ([]string, error) {
	return s.supported(ctx, nil)
}

// SupportedWithFixedRate returns supported currencies with ?fixed_rate=true or false.
func (s *CurrenciesService) SupportedWithFixedRate(ctx context.Context, fixedRate bool) ([]string, error) {
	return s.supported(ctx, &fixedRate)
}

func (s *CurrenciesService) supported(ctx context.Context, fixedRate *bool) ([]string, error) {
	path := "/v1/currencies"
	if fixedRate != nil {
		if *fixedRate {
			path += "?fixed_rate=true"
		} else {
			path += "?fixed_rate=false"
		}
	}
	var out models.CurrenciesResponse
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// Available returns available currencies (GET /v1/currencies/available).
func (s *CurrenciesService) Available(ctx context.Context) ([]string, error) {
	var out models.CurrenciesResponse
	err := s.c.do.Do(ctx, "GET", "/v1/currencies/available", nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// FullCurrencies returns detailed currency list (GET /v1/full-currencies).
func (s *CurrenciesService) FullCurrencies(ctx context.Context) (*models.FullCurrenciesResponse, error) {
	var out models.FullCurrenciesResponse
	err := s.c.do.Do(ctx, "GET", "/v1/full-currencies", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// MerchantCoins returns coins set as available in merchant "coins settings" (GET /v1/merchant/coins).
func (s *CurrenciesService) MerchantCoins(ctx context.Context) ([]string, error) {
	var out models.CurrenciesResponse
	err := s.c.do.Do(ctx, "GET", "/v1/merchant/coins", nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// EstimatesService handles price/amount estimates and fiat conversion.
type EstimatesService struct{ c *Client }

// Estimate returns an estimate for the given amount and currency pair (e.g. 10 USD -> BTC).
func (s *EstimatesService) Estimate(ctx context.Context, amount float64, currencyFrom, currencyTo string) (*models.Estimate, error) {
	path := "/v1/estimate?" + url.Values{
		"amount":         {formatFloat(amount)},
		"currency_from":  {currencyFrom},
		"currency_to":    {currencyTo},
	}.Encode()
	var out models.Estimate
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// EstimateByPrice returns estimate when paying a given fiat amount (e.g. price in USD) in a crypto.
func (s *EstimatesService) EstimateByPrice(ctx context.Context, priceAmount float64, priceCurrency, payCurrency string) (*models.Estimate, error) {
	return s.Estimate(ctx, priceAmount, priceCurrency, payCurrency)
}

// MinAmount returns the minimum payment amount for a currency pair.
func (s *EstimatesService) MinAmount(ctx context.Context, currencyFrom, currencyTo string) (float64, error) {
	resp, err := s.MinAmountEx(ctx, &models.MinAmountParams{CurrencyFrom: currencyFrom, CurrencyTo: currencyTo})
	if err != nil {
		return 0, err
	}
	return resp.MinAmount, nil
}

// MinAmountEx returns the full min-amount response with optional fiat_equivalent, is_fixed_rate, is_fee_paid_by_user. Params must be non-nil.
func (s *EstimatesService) MinAmountEx(ctx context.Context, params *models.MinAmountParams) (*models.MinAmountResponse, error) {
	if params == nil {
		panic("nowpay: MinAmountParams is required")
	}
	v := url.Values{}
	v.Set("currency_from", params.CurrencyFrom)
	v.Set("currency_to", params.CurrencyTo)
	if params.FiatEquivalent != "" {
		v.Set("fiat_equivalent", params.FiatEquivalent)
	}
	if params.IsFixedRate {
		v.Set("is_fixed_rate", "true")
	}
	if params.IsFeePaidByUser {
		v.Set("is_fee_paid_by_user", "true")
	}
	path := "/v1/min-amount?" + v.Encode()
	var out models.MinAmountResponse
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// PayoutsService handles payout operations.
type PayoutsService struct{ c *Client }

// Create creates a payout.
func (s *PayoutsService) Create(ctx context.Context, req *models.CreatePayoutRequest) (*models.Payout, error) {
	var out models.Payout
	err := s.c.do.Do(ctx, "POST", "/v1/payout", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a payout by withdrawal ID.
func (s *PayoutsService) Get(ctx context.Context, withdrawalID string) (*models.Payout, error) {
	var out models.Payout
	err := s.c.do.Do(ctx, "GET", "/v1/payout/"+withdrawalID, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// BatchCreate creates multiple payouts in one request (mass payouts API).
func (s *PayoutsService) BatchCreate(ctx context.Context, req *models.BatchPayoutRequest) (*models.BatchPayoutResponse, error) {
	var out models.BatchPayoutResponse
	err := s.c.do.Do(ctx, "POST", "/v1/payout/batch", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ValidateAddress validates a payout address (POST /v1/payout/validate-address).
func (s *PayoutsService) ValidateAddress(ctx context.Context, req *models.ValidateAddressRequest) error {
	return s.c.do.Do(ctx, "POST", "/v1/payout/validate-address", req, nil)
}

// List returns a list of payouts (GET /v1/payout?batch_id=...&status=...).
func (s *PayoutsService) List(ctx context.Context, params *models.ListPayoutsParams) ([]models.Payout, error) {
	path := "/v1/payout"
	if params != nil {
		if q := encodeListPayoutsParams(params); q != "" {
			path += "?" + q
		}
	}
	var out []models.Payout
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MinAmountForWithdrawal returns minimal withdrawal amount for a coin (GET /v1/payout-withdrawal/min-amount/:coin).
func (s *PayoutsService) MinAmountForWithdrawal(ctx context.Context, coin string) (float64, error) {
	var out struct {
		MinAmount float64 `json:"min_amount"`
	}
	err := s.c.do.Do(ctx, "GET", "/v1/payout-withdrawal/min-amount/"+url.PathEscape(coin), nil, &out)
	if err != nil {
		return 0, err
	}
	return out.MinAmount, nil
}

// Fee returns withdrawal fee estimate (GET /v1/payout/fee?currency=...&amount=...).
func (s *PayoutsService) Fee(ctx context.Context, currency string, amount float64) (float64, error) {
	path := "/v1/payout/fee?" + url.Values{"currency": {currency}, "amount": {formatFloat(amount)}}.Encode()
	var out struct {
		Fee float64 `json:"fee,omitempty"`
	}
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return 0, err
	}
	return out.Fee, nil
}

// Cancel cancels a scheduled payout (POST /v1/payout/:w_id/cancel).
func (s *PayoutsService) Cancel(ctx context.Context, withdrawalID string) error {
	return s.c.do.Do(ctx, "POST", "/v1/payout/"+url.PathEscape(withdrawalID)+"/cancel", nil, nil)
}

// Verify verifies a batch payout (GET /v1/payout/:batch-withdrawal-id/verify).
func (s *PayoutsService) Verify(ctx context.Context, batchWithdrawalID string) (*models.Payout, error) {
	var out models.Payout
	err := s.c.do.Do(ctx, "GET", "/v1/payout/"+url.PathEscape(batchWithdrawalID)+"/verify", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// SubscriptionsService handles subscription operations.
type SubscriptionsService struct{ c *Client }

// Create creates a subscription.
func (s *SubscriptionsService) Create(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	var out models.Subscription
	err := s.c.do.Do(ctx, "POST", "/v1/subscription", req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a subscription by ID.
func (s *SubscriptionsService) Get(ctx context.Context, subscriptionID string) (*models.Subscription, error) {
	var out models.Subscription
	err := s.c.do.Do(ctx, "GET", "/v1/subscription/"+subscriptionID, nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels a subscription.
func (s *SubscriptionsService) Cancel(ctx context.Context, subscriptionID string) error {
	return s.c.do.Do(ctx, "DELETE", "/v1/subscription/"+subscriptionID, nil, nil)
}

// Update updates a subscription.
func (s *SubscriptionsService) Update(ctx context.Context, subscriptionID string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	var out models.Subscription
	err := s.c.do.Do(ctx, "PATCH", "/v1/subscription/"+subscriptionID, req, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func formatInt64(n int64) string { return strconv.FormatInt(n, 10) }

func formatFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func encodeListPaymentsParams(p *models.ListPaymentsParams) string {
	v := url.Values{}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		v.Set("page", strconv.Itoa(p.Page))
	}
	if p.SortBy != "" {
		v.Set("sortBy", p.SortBy)
	}
	if p.OrderBy != "" {
		v.Set("orderBy", p.OrderBy)
	}
	if p.DateFrom != "" {
		v.Set("dateFrom", p.DateFrom)
	}
	if p.DateTo != "" {
		v.Set("dateTo", p.DateTo)
	}
	if p.InvoiceID != "" {
		v.Set("invoiceId", p.InvoiceID)
	}
	return v.Encode()
}

func encodeListPayoutsParams(p *models.ListPayoutsParams) string {
	v := url.Values{}
	if p.BatchID != "" {
		v.Set("batch_id", p.BatchID)
	}
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	if p.OrderBy != "" {
		v.Set("order_by", p.OrderBy)
	}
	if p.Order != "" {
		v.Set("order", p.Order)
	}
	if p.DateFrom != "" {
		v.Set("date_from", p.DateFrom)
	}
	if p.DateTo != "" {
		v.Set("date_to", p.DateTo)
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		v.Set("page", strconv.Itoa(p.Page))
	}
	return v.Encode()
}
