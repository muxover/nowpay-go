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

// Config configures the NowPay client.
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	HTTPClient *http.Client
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
	do := internal.NewClient(baseURL, cfg.APIKey, timeout, cfg.HTTPClient)
	c := &Client{do: do}
	c.Payments = &PaymentsService{c: c}
	c.Invoices = &InvoicesService{c: c}
	c.Currencies = &CurrenciesService{c: c}
	c.Estimates = &EstimatesService{c: c}
	c.Payouts = &PayoutsService{c: c}
	c.Subscriptions = &SubscriptionsService{c: c}
	return c
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

// Supported returns the list of supported currencies (simple list).
func (s *CurrenciesService) Supported(ctx context.Context) ([]string, error) {
	var out models.CurrenciesResponse
	err := s.c.do.Do(ctx, "GET", "/v1/currencies", nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// Available returns available currencies (full list with metadata when API supports it).
func (s *CurrenciesService) Available(ctx context.Context) ([]string, error) {
	var out models.CurrenciesResponse
	err := s.c.do.Do(ctx, "GET", "/v1/currencies/available", nil, &out)
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

// MinAmount returns the minimum payment amount for a currency.
func (s *EstimatesService) MinAmount(ctx context.Context, currencyFrom, currencyTo string) (float64, error) {
	path := "/v1/min-amount?" + url.Values{
		"currency_from": {currencyFrom},
		"currency_to":   {currencyTo},
	}.Encode()
	var out models.MinAmountResponse
	err := s.c.do.Do(ctx, "GET", path, nil, &out)
	if err != nil {
		return 0, err
	}
	return out.MinAmount, nil
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
	return v.Encode()
}
