package nowpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ErrSubscriptionNotFound = internal.ErrSubscriptionNotFound
	ErrRateLimited          = internal.ErrRateLimited
	ErrBadRequest           = internal.ErrBadRequest
	ErrValidation           = internal.ErrValidation
	ErrServerError          = internal.ErrServerError
	ErrUnauthorized         = internal.ErrUnauthorized
	ErrForbidden            = internal.ErrForbidden
	ErrNotFound             = internal.ErrNotFound
)

// APIError is the error returned for any 4xx or 5xx response. Use errors.As
// to read the status code and the NOWPayments error code and message.
type APIError = internal.APIError

// IsAPIError reports whether err is or wraps an *APIError.
func IsAPIError(err error) bool { return internal.IsAPIError(err) }

// DefaultBaseURL already includes the /v1 version segment; endpoint paths are
// joined relative to it (e.g. "/payment").
const DefaultBaseURL = "https://api.nowpayments.io/v1"

// SandboxBaseURL is the NOWPayments sandbox. It needs a sandbox account and API key.
const SandboxBaseURL = "https://api-sandbox.nowpayments.io/v1"

// RetryConfig configures retry behavior for 5xx and 429 responses. Zero value means no retries.
type RetryConfig struct {
	MaxRetries     int           // max attempts (0 = no retries, 3 recommended)
	InitialBackoff time.Duration // first backoff (e.g. 1*time.Second)
	MaxBackoff     time.Duration // cap (e.g. 30*time.Second)
}

// Config configures a Client. Only APIKey is required.
type Config struct {
	APIKey     string
	Token      string        // optional JWT (from Auth); required for List payments, payouts, conversions, custody
	BaseURL    string        // defaults to DefaultBaseURL
	Timeout    time.Duration // defaults to 30s; ignored when HTTPClient is set
	HTTPClient *http.Client
	Retry      RetryConfig // optional retry for 429, and for 5xx on idempotent requests
}

// Client is a NOWPayments API client. It is safe for concurrent use.
type Client struct {
	Payments      *PaymentsService
	Invoices      *InvoicesService
	Currencies    *CurrenciesService
	Estimates     *EstimatesService
	Payouts       *PayoutsService
	Subscriptions *SubscriptionsService
	Conversions   *ConversionsService
	Customers     *CustomersService
	FiatPayouts   *FiatPayoutsService
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
	c.Conversions = &ConversionsService{c: c}
	c.Customers = &CustomersService{c: c}
	c.FiatPayouts = &FiatPayoutsService{c: c}
	return c
}

// notFound lets a 404 match ErrNotFound and the resource-specific sentinel.
func notFound(err, target error) error {
	var ae *APIError
	if errors.As(err, &ae) && ae.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %w", target, err)
	}
	return err
}

// getResult decodes a {"result": ...} envelope used by many NOWPayments endpoints.
func getResult[T any](ctx context.Context, c *Client, method, path string, body any) (T, error) {
	var env struct {
		Result T `json:"result"`
	}
	if err := c.do.Do(ctx, method, path, body, &env); err != nil {
		var zero T
		return zero, err
	}
	return env.Result, nil
}

// Status checks API availability (GET /status).
func (c *Client) Status(ctx context.Context) (*models.StatusResponse, error) {
	var out models.StatusResponse
	if err := c.do.Do(ctx, "GET", "/status", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Auth obtains a JWT token (POST /auth). Required for payouts, conversions, custody, and list payments on some accounts.
func (c *Client) Auth(ctx context.Context, email, password string) (*models.AuthResponse, error) {
	var out models.AuthResponse
	if err := c.do.Do(ctx, "POST", "/auth", &models.AuthRequest{Email: email, Password: password}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance returns balance per currency (GET /balance).
func (c *Client) Balance(ctx context.Context) (*models.BalanceResponse, error) {
	var out models.BalanceResponse
	if err := c.do.Do(ctx, "GET", "/balance", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PaymentsService handles the standard payments API.
type PaymentsService struct{ c *Client }

// Create creates a payment (POST /payment).
func (s *PaymentsService) Create(ctx context.Context, req *models.CreatePaymentRequest) (*models.Payment, error) {
	var out models.Payment
	if err := s.c.do.Do(ctx, "POST", "/payment", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a payment's current state (GET /payment/:id).
func (s *PaymentsService) Get(ctx context.Context, paymentID int64) (*models.Payment, error) {
	var out models.Payment
	if err := s.c.do.Do(ctx, "GET", "/payment/"+formatInt64(paymentID), nil, &out); err != nil {
		return nil, notFound(err, ErrPaymentNotFound)
	}
	return &out, nil
}

// List returns payments, newest first unless SortBy/OrderBy say otherwise.
// Pass nil for the defaults. Requires a JWT on most accounts.
func (s *PaymentsService) List(ctx context.Context, params *models.ListPaymentsParams) ([]models.Payment, error) {
	if params == nil {
		params = &models.ListPaymentsParams{}
	}
	path := "/payment/?" + encodeListPaymentsParams(params)
	var out models.ListPaymentsResponse
	if err := s.c.do.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// CreateFromInvoice creates a payment from an invoice (POST /invoice-payment).
func (s *PaymentsService) CreateFromInvoice(ctx context.Context, req *models.CreateInvoicePaymentRequest) (*models.Payment, error) {
	var out models.Payment
	if err := s.c.do.Do(ctx, "POST", "/invoice-payment", req, &out); err != nil {
		return nil, notFound(err, ErrInvoiceNotFound)
	}
	return &out, nil
}

// UpdateMerchantEstimate gets/updates the payment estimate (POST /payment/:id/update-merchant-estimate).
func (s *PaymentsService) UpdateMerchantEstimate(ctx context.Context, paymentID int64) (*models.UpdateMerchantEstimateResponse, error) {
	var out models.UpdateMerchantEstimateResponse
	if err := s.c.do.Do(ctx, "POST", "/payment/"+formatInt64(paymentID)+"/update-merchant-estimate", nil, &out); err != nil {
		return nil, notFound(err, ErrPaymentNotFound)
	}
	return &out, nil
}

// InvoicesService handles invoices.
type InvoicesService struct{ c *Client }

// Create creates a hosted invoice (POST /invoice).
func (s *InvoicesService) Create(ctx context.Context, req *models.CreateInvoiceRequest) (*models.Invoice, error) {
	var out models.Invoice
	if err := s.c.do.Do(ctx, "POST", "/invoice", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CurrenciesService handles currency listings.
type CurrenciesService struct{ c *Client }

// Supported returns the list of supported currencies (GET /currencies).
func (s *CurrenciesService) Supported(ctx context.Context) ([]string, error) {
	return s.supported(ctx, nil)
}

// SupportedWithFixedRate returns supported currencies with ?fixed_rate=true or false.
func (s *CurrenciesService) SupportedWithFixedRate(ctx context.Context, fixedRate bool) ([]string, error) {
	return s.supported(ctx, &fixedRate)
}

func (s *CurrenciesService) supported(ctx context.Context, fixedRate *bool) ([]string, error) {
	path := "/currencies"
	if fixedRate != nil {
		if *fixedRate {
			path += "?fixed_rate=true"
		} else {
			path += "?fixed_rate=false"
		}
	}
	var out models.CurrenciesResponse
	if err := s.c.do.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// FullCurrencies returns detailed currency list (GET /full-currencies).
func (s *CurrenciesService) FullCurrencies(ctx context.Context) (*models.FullCurrenciesResponse, error) {
	var out models.FullCurrenciesResponse
	if err := s.c.do.Do(ctx, "GET", "/full-currencies", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MerchantCoins returns coins set as available in merchant "coins settings" (GET /merchant/coins).
func (s *CurrenciesService) MerchantCoins(ctx context.Context) ([]string, error) {
	var out models.CurrenciesResponse
	if err := s.c.do.Do(ctx, "GET", "/merchant/coins", nil, &out); err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// EstimatesService handles price/amount estimates and fiat conversion.
type EstimatesService struct{ c *Client }

// Estimate returns an estimate for the given amount and currency pair (e.g. 10 USD -> BTC).
func (s *EstimatesService) Estimate(ctx context.Context, amount float64, currencyFrom, currencyTo string) (*models.Estimate, error) {
	path := "/estimate?" + url.Values{
		"amount":        {formatFloat(amount)},
		"currency_from": {currencyFrom},
		"currency_to":   {currencyTo},
	}.Encode()
	var out models.Estimate
	if err := s.c.do.Do(ctx, "GET", path, nil, &out); err != nil {
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

// MinAmountEx returns the full min-amount response with optional fiat_equivalent, is_fixed_rate, is_fee_paid_by_user.
func (s *EstimatesService) MinAmountEx(ctx context.Context, params *models.MinAmountParams) (*models.MinAmountResponse, error) {
	if params == nil {
		return nil, errors.New("nowpay: MinAmountEx: params are required")
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
	path := "/min-amount?" + v.Encode()
	var out models.MinAmountResponse
	if err := s.c.do.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayoutsService lists mass payouts and checks addresses, minimums, and fees.
// It requires a JWT (set Config.Token).
type PayoutsService struct{ c *Client }

// List returns a list of payouts (GET /payout).
func (s *PayoutsService) List(ctx context.Context, params *models.ListPayoutsParams) ([]models.Payout, error) {
	path := "/payout"
	if params != nil {
		if q := encodeListPayoutsParams(params); q != "" {
			path += "?" + q
		}
	}
	var out models.ListPayoutsResponse
	if err := s.c.do.Do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return out.Payouts, nil
}

// ValidateAddress validates a payout address (POST /payout/validate-address). Returns nil if valid.
func (s *PayoutsService) ValidateAddress(ctx context.Context, req *models.ValidateAddressRequest) error {
	return s.c.do.Do(ctx, "POST", "/payout/validate-address", req, nil)
}

// MinAmountForWithdrawal returns minimal withdrawal amount for a coin (GET /payout-withdrawal/min-amount/:coin).
func (s *PayoutsService) MinAmountForWithdrawal(ctx context.Context, coin string) (float64, error) {
	var out struct {
		Result    models.FlexFloat `json:"result"`
		MinAmount models.FlexFloat `json:"min_amount"`
	}
	if err := s.c.do.Do(ctx, "GET", "/payout-withdrawal/min-amount/"+url.PathEscape(coin), nil, &out); err != nil {
		return 0, err
	}
	if out.Result != 0 {
		return out.Result.Float64(), nil
	}
	return out.MinAmount.Float64(), nil
}

// Fee returns withdrawal fee estimate (GET /payout/fee?currency=...&amount=...).
func (s *PayoutsService) Fee(ctx context.Context, currency string, amount float64) (float64, error) {
	path := "/payout/fee?" + url.Values{"currency": {currency}, "amount": {formatFloat(amount)}}.Encode()
	var out struct {
		Fee models.FlexFloat `json:"fee"`
	}
	if err := s.c.do.Do(ctx, "GET", path, nil, &out); err != nil {
		return 0, err
	}
	return out.Fee.Float64(), nil
}

// SubscriptionsService handles recurring payments: plans and email/customer subscriptions.
type SubscriptionsService struct{ c *Client }

// CreatePlan creates a subscription plan (POST /subscriptions/plans).
func (s *SubscriptionsService) CreatePlan(ctx context.Context, req *models.CreatePlanRequest) (*models.SubscriptionPlan, error) {
	out, err := getResult[models.SubscriptionPlan](ctx, s.c, "POST", "/subscriptions/plans", req)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdatePlan updates a subscription plan (PATCH /subscriptions/plans/:id).
func (s *SubscriptionsService) UpdatePlan(ctx context.Context, planID string, req *models.UpdatePlanRequest) (*models.SubscriptionPlan, error) {
	out, err := getResult[models.SubscriptionPlan](ctx, s.c, "PATCH", "/subscriptions/plans/"+url.PathEscape(planID), req)
	if err != nil {
		return nil, notFound(err, ErrSubscriptionNotFound)
	}
	return &out, nil
}

// GetPlan returns a subscription plan (GET /subscriptions/plans/:id).
func (s *SubscriptionsService) GetPlan(ctx context.Context, planID string) (*models.SubscriptionPlan, error) {
	out, err := getResult[models.SubscriptionPlan](ctx, s.c, "GET", "/subscriptions/plans/"+url.PathEscape(planID), nil)
	if err != nil {
		return nil, notFound(err, ErrSubscriptionNotFound)
	}
	return &out, nil
}

// ListPlans returns subscription plans (GET /subscriptions/plans).
func (s *SubscriptionsService) ListPlans(ctx context.Context, params *models.ListPlansParams) ([]models.SubscriptionPlan, error) {
	path := "/subscriptions/plans"
	if params != nil {
		v := url.Values{}
		if params.Limit > 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			v.Set("offset", strconv.Itoa(params.Offset))
		}
		if q := v.Encode(); q != "" {
			path += "?" + q
		}
	}
	return getResult[[]models.SubscriptionPlan](ctx, s.c, "GET", path, nil)
}

// Create creates an email/customer recurring payment (POST /subscriptions).
// The API returns the new subscription as {"result": [...]}; the docs show
// {"result": {...}}. Both are accepted.
func (s *SubscriptionsService) Create(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	raw, err := getResult[json.RawMessage](ctx, s.c, "POST", "/subscriptions", req)
	if err != nil {
		return nil, err
	}
	var list []models.Subscription
	if err := json.Unmarshal(raw, &list); err == nil {
		if len(list) == 0 {
			return nil, errors.New("nowpay: POST /subscriptions returned no subscription")
		}
		return &list[0], nil
	}
	var out models.Subscription
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("nowpay: decode POST /subscriptions response: %w", err)
	}
	return &out, nil
}

// Get returns a recurring payment (GET /subscriptions/:id).
func (s *SubscriptionsService) Get(ctx context.Context, subscriptionID string) (*models.Subscription, error) {
	out, err := getResult[models.Subscription](ctx, s.c, "GET", "/subscriptions/"+url.PathEscape(subscriptionID), nil)
	if err != nil {
		return nil, notFound(err, ErrSubscriptionNotFound)
	}
	return &out, nil
}

// List returns recurring payments (GET /subscriptions).
func (s *SubscriptionsService) List(ctx context.Context, params *models.ListSubscriptionsParams) ([]models.Subscription, error) {
	path := "/subscriptions"
	if params != nil {
		v := url.Values{}
		if params.Status != "" {
			v.Set("status", params.Status)
		}
		if params.SubscriptionPlanID > 0 {
			v.Set("subscription_plan_id", formatInt64(params.SubscriptionPlanID))
		}
		if params.IsActive != nil {
			v.Set("is_active", strconv.FormatBool(*params.IsActive))
		}
		if params.Limit > 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			v.Set("offset", strconv.Itoa(params.Offset))
		}
		if q := v.Encode(); q != "" {
			path += "?" + q
		}
	}
	return getResult[[]models.Subscription](ctx, s.c, "GET", path, nil)
}

// Delete deletes a recurring payment (DELETE /subscriptions/:id).
func (s *SubscriptionsService) Delete(ctx context.Context, subscriptionID string) error {
	return notFound(s.c.do.Do(ctx, "DELETE", "/subscriptions/"+url.PathEscape(subscriptionID), nil, nil), ErrSubscriptionNotFound)
}

// ConversionsService lists custody currency conversions (JWT required).
type ConversionsService struct{ c *Client }

// List returns conversions (GET /conversion).
func (s *ConversionsService) List(ctx context.Context, params *models.ListConversionsParams) ([]models.Conversion, error) {
	path := "/conversion"
	if params != nil {
		v := url.Values{}
		if params.ID != "" {
			v.Set("id", params.ID)
		}
		if params.Status != "" {
			v.Set("status", params.Status)
		}
		if params.FromCurrency != "" {
			v.Set("from_currency", params.FromCurrency)
		}
		if params.ToCurrency != "" {
			v.Set("to_currency", params.ToCurrency)
		}
		if params.CreatedFrom != "" {
			v.Set("created_at_from", params.CreatedFrom)
		}
		if params.CreatedTo != "" {
			v.Set("created_at_to", params.CreatedTo)
		}
		if params.Limit > 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			v.Set("offset", strconv.Itoa(params.Offset))
		}
		if params.Order != "" {
			v.Set("order", params.Order)
		}
		if q := v.Encode(); q != "" {
			path += "?" + q
		}
	}
	return getResult[[]models.Conversion](ctx, s.c, "GET", path, nil)
}

// CustomersService handles custody / customer management (sub-partner) endpoints (JWT required).
type CustomersService struct{ c *Client }

// Create creates a new customer / sub-account (POST /sub-partner/balance).
func (s *CustomersService) Create(ctx context.Context, req *models.CreateCustomerRequest) (*models.Customer, error) {
	out, err := getResult[models.Customer](ctx, s.c, "POST", "/sub-partner/balance", req)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Balance returns a customer's balances (GET /sub-partner/balance/:id).
func (s *CustomersService) Balance(ctx context.Context, customerID string) (*models.CustomerBalance, error) {
	out, err := getResult[models.CustomerBalance](ctx, s.c, "GET", "/sub-partner/balance/"+url.PathEscape(customerID), nil)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns customers / sub-accounts (GET /sub-partner).
func (s *CustomersService) List(ctx context.Context, params *models.ListCustomersParams) ([]models.Customer, error) {
	path := "/sub-partner"
	if params != nil {
		v := url.Values{}
		if params.ID != "" {
			v.Set("id", params.ID)
		}
		if params.Offset > 0 {
			v.Set("offset", strconv.Itoa(params.Offset))
		}
		if params.Limit > 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Order != "" {
			v.Set("order", params.Order)
		}
		if q := v.Encode(); q != "" {
			path += "?" + q
		}
	}
	return getResult[[]models.Customer](ctx, s.c, "GET", path, nil)
}

// Transfers returns custody transfers (GET /sub-partner/transfers).
func (s *CustomersService) Transfers(ctx context.Context, params *models.ListTransfersParams) ([]models.CustodyTransfer, error) {
	path := "/sub-partner/transfers"
	if params != nil {
		v := url.Values{}
		if params.ID != "" {
			v.Set("id", params.ID)
		}
		if params.Status != "" {
			v.Set("status", params.Status)
		}
		if params.Limit > 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Offset > 0 {
			v.Set("offset", strconv.Itoa(params.Offset))
		}
		if params.Order != "" {
			v.Set("order", params.Order)
		}
		if q := v.Encode(); q != "" {
			path += "?" + q
		}
	}
	return getResult[[]models.CustodyTransfer](ctx, s.c, "GET", path, nil)
}

// DepositWithPayment creates a deposit-with-payment for a customer (POST /sub-partner/payment).
func (s *CustomersService) DepositWithPayment(ctx context.Context, req *models.CustodyDepositRequest) (*models.Payment, error) {
	out, err := getResult[models.Payment](ctx, s.c, "POST", "/sub-partner/payment", req)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Payments returns customer (sub-account) payments (GET /sub-partner/payments).
func (s *CustomersService) Payments(ctx context.Context) ([]models.Payment, error) {
	return getResult[[]models.Payment](ctx, s.c, "GET", "/sub-partner/payments", nil)
}

// FiatPayoutsService reads fiat (bank) payout providers, currencies, accounts,
// and payouts (JWT required).
type FiatPayoutsService struct{ c *Client }

// Providers returns available fiat-payout providers (GET /fiat-payouts/providers).
func (s *FiatPayoutsService) Providers(ctx context.Context) ([]models.FiatProvider, error) {
	return getResult[[]models.FiatProvider](ctx, s.c, "GET", "/fiat-payouts/providers", nil)
}

// FiatCurrencies returns the fiat currencies a provider supports
// (GET /fiat-payouts/fiat-currencies). The API requires a provider code; see Providers.
func (s *FiatPayoutsService) FiatCurrencies(ctx context.Context, provider string) ([]models.FiatCurrency, error) {
	return getResult[[]models.FiatCurrency](ctx, s.c, "GET", "/fiat-payouts/fiat-currencies"+providerQuery(provider), nil)
}

func providerQuery(provider string) string {
	if provider == "" {
		return ""
	}
	return "?" + url.Values{"provider": {provider}}.Encode()
}

// CryptoCurrencies returns supported crypto currencies (GET /fiat-payouts/crypto-currencies). provider and currency are optional filters.
func (s *FiatPayoutsService) CryptoCurrencies(ctx context.Context, provider, currency string) ([]models.FiatCryptoCurrency, error) {
	path := "/fiat-payouts/crypto-currencies"
	v := url.Values{}
	if provider != "" {
		v.Set("provider", provider)
	}
	if currency != "" {
		v.Set("currency", currency)
	}
	if q := v.Encode(); q != "" {
		path += "?" + q
	}
	return getResult[[]models.FiatCryptoCurrency](ctx, s.c, "GET", path, nil)
}

// PaymentMethods returns available payment methods (GET /fiat-payouts/payment-methods). provider and currency are optional filters.
func (s *FiatPayoutsService) PaymentMethods(ctx context.Context, provider, currency string) ([]models.FiatPaymentMethod, error) {
	path := "/fiat-payouts/payment-methods"
	v := url.Values{}
	if provider != "" {
		v.Set("provider", provider)
	}
	if currency != "" {
		v.Set("currency", currency)
	}
	if q := v.Encode(); q != "" {
		path += "?" + q
	}
	return getResult[[]models.FiatPaymentMethod](ctx, s.c, "GET", path, nil)
}

// Accounts returns fiat payout accounts (GET /fiat-payouts/accounts). An empty
// provider returns the accounts of every provider.
func (s *FiatPayoutsService) Accounts(ctx context.Context, provider string) (map[string]any, error) {
	return getResult[map[string]any](ctx, s.c, "GET", "/fiat-payouts/accounts"+providerQuery(provider), nil)
}

// List returns fiat payouts (GET /fiat-payouts).
func (s *FiatPayoutsService) List(ctx context.Context, params *models.ListFiatPayoutsParams) ([]models.FiatPayout, error) {
	path := "/fiat-payouts"
	if params != nil {
		v := url.Values{}
		if params.ID != "" {
			v.Set("id", params.ID)
		}
		if params.Provider != "" {
			v.Set("provider", params.Provider)
		}
		if params.RequestID != "" {
			v.Set("requestId", params.RequestID)
		}
		if params.FiatCurrency != "" {
			v.Set("fiatCurrency", params.FiatCurrency)
		}
		if params.CryptoCurrency != "" {
			v.Set("cryptoCurrency", params.CryptoCurrency)
		}
		if params.Status != "" {
			v.Set("status", params.Status)
		}
		if params.Limit > 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Page > 0 {
			v.Set("page", strconv.Itoa(params.Page))
		}
		if params.OrderBy != "" {
			v.Set("orderBy", params.OrderBy)
		}
		if params.SortBy != "" {
			v.Set("sortBy", params.SortBy)
		}
		if params.DateFrom != "" {
			v.Set("dateFrom", params.DateFrom)
		}
		if params.DateTo != "" {
			v.Set("dateTo", params.DateTo)
		}
		if q := v.Encode(); q != "" {
			path += "?" + q
		}
	}
	rows, err := getResult[models.FiatPayoutRows](ctx, s.c, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	return rows.Rows, nil
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
	// The API answers 500 "Invalid order_by value" when these are missing.
	sortBy, orderBy := p.SortBy, p.OrderBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	if orderBy == "" {
		orderBy = "desc"
	}
	v.Set("sortBy", sortBy)
	v.Set("orderBy", orderBy)
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
