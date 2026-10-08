package nowpay

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/muxover/nowpay-go/models"
)

// The fixtures in testdata are the example responses from the official
// NOWPayments Postman collection.
func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

type recorded struct {
	method string
	uri    string
	body   string
	header http.Header
}

// serve starts a server that answers every request with status and body,
// and records the last request.
func serve(t *testing.T, status int, body []byte) (*Client, *recorded) {
	t.Helper()
	rec := &recorded{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		rec.method, rec.uri, rec.body, rec.header = r.Method, r.URL.RequestURI(), string(b), r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return NewClient(Config{APIKey: "test-key", Token: "jwt", BaseURL: srv.URL + "/v1"}), rec
}

func jsonEqual(t *testing.T, got, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal([]byte(got), &g); err != nil {
		t.Fatalf("request body is not JSON: %q", got)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("request body\n got: %s\nwant: %s", got, want)
	}
}

func TestEndpoints(t *testing.T) {
	ctx := context.Background()
	yes := true

	tests := []struct {
		name    string
		fixture string
		method  string
		uri     string
		reqBody string
		call    func(t *testing.T, c *Client) error
	}{
		{"Status", "status", "GET", "/v1/status", "", func(t *testing.T, c *Client) error {
			out, err := c.Status(ctx)
			if err == nil && out.Message != "OK" {
				t.Errorf("message = %q", out.Message)
			}
			return err
		}},
		{"Auth", "auth", "POST", "/v1/auth", `{"email":"a@b.c","password":"pw"}`, func(t *testing.T, c *Client) error {
			out, err := c.Auth(ctx, "a@b.c", "pw")
			if err == nil && out.Token == "" {
				t.Error("empty token")
			}
			return err
		}},
		{"Balance", "balance", "GET", "/v1/balance", "", func(t *testing.T, c *Client) error {
			out, err := c.Balance(ctx)
			if err == nil && (len(out.Currencies) != 3 || out.Currencies["eth"].Amount == 0) {
				t.Errorf("balances = %+v", out.Currencies)
			}
			return err
		}},
		{"Payments.Create", "payment_create", "POST", "/v1/payment", `{"price_amount":3999.5,"price_currency":"usd","pay_currency":"btc","order_id":"RGDBP-21314"}`, func(t *testing.T, c *Client) error {
			out, err := c.Payments.Create(ctx, &models.CreatePaymentRequest{PriceAmount: 3999.5, PriceCurrency: "usd", PayCurrency: "btc", OrderID: "RGDBP-21314"})
			if err == nil && (out.PaymentID != 5745459419 || out.PurchaseID != 5837122679 || out.PayAmount != 0.17070286) {
				t.Errorf("payment = %+v", out)
			}
			return err
		}},
		{"Payments.Get", "payment_get", "GET", "/v1/payment/6249365965", "", func(t *testing.T, c *Client) error {
			out, err := c.Payments.Get(ctx, 6249365965)
			if err == nil && (out.PaymentStatus != "finished" || len(out.PaymentExtraIDs) != 1 || out.OutcomeAmount != 11.8405) {
				t.Errorf("payment = %+v", out)
			}
			return err
		}},
		{"Payments.List", "payment_list", "GET", "/v1/payment/?invoiceId=6200264890&limit=10&orderBy=asc&sortBy=created_at", "", func(t *testing.T, c *Client) error {
			out, err := c.Payments.List(ctx, &models.ListPaymentsParams{Limit: 10, SortBy: "created_at", OrderBy: "asc", InvoiceID: "6200264890"})
			if err == nil && (len(out) != 2 || out[0].InvoiceID != 6200264890) {
				t.Errorf("payments = %+v", out)
			}
			return err
		}},
		{"Payments.List defaults", "payment_list", "GET", "/v1/payment/?orderBy=desc&sortBy=created_at", "", func(t *testing.T, c *Client) error {
			_, err := c.Payments.List(ctx, nil)
			return err
		}},
		{"Payments.CreateFromInvoice", "payment_create", "POST", "/v1/invoice-payment", `{"iid":4522625843,"pay_currency":"btc"}`, func(t *testing.T, c *Client) error {
			_, err := c.Payments.CreateFromInvoice(ctx, &models.CreateInvoicePaymentRequest{IID: 4522625843, PayCurrency: "btc"})
			return err
		}},
		{"Payments.UpdateMerchantEstimate", "update_estimate", "POST", "/v1/payment/4455667788/update-merchant-estimate", "", func(t *testing.T, c *Client) error {
			out, err := c.Payments.UpdateMerchantEstimate(ctx, 4455667788)
			if err == nil && out.TokenID != 5566778899 {
				t.Errorf("estimate = %+v", out)
			}
			return err
		}},
		{"Invoices.Create", "invoice_create", "POST", "/v1/invoice", `{"price_amount":1000,"price_currency":"usd","order_id":"RGDBP-21314"}`, func(t *testing.T, c *Client) error {
			out, err := c.Invoices.Create(ctx, &models.CreateInvoiceRequest{PriceAmount: 1000, PriceCurrency: "usd", OrderID: "RGDBP-21314"})
			if err == nil && (out.ID != 4522625843 || out.PriceAmount != 1000 || out.InvoiceURL == "") {
				t.Errorf("invoice = %+v", out)
			}
			return err
		}},
		{"Currencies.Supported", "currencies", "GET", "/v1/currencies", "", func(t *testing.T, c *Client) error {
			out, err := c.Currencies.Supported(ctx)
			if err == nil && len(out) != 6 {
				t.Errorf("currencies = %v", out)
			}
			return err
		}},
		{"Currencies.SupportedWithFixedRate", "currencies", "GET", "/v1/currencies?fixed_rate=true", "", func(t *testing.T, c *Client) error {
			_, err := c.Currencies.SupportedWithFixedRate(ctx, true)
			return err
		}},
		{"Currencies.FullCurrencies", "full_currencies", "GET", "/v1/full-currencies", "", func(t *testing.T, c *Client) error {
			out, err := c.Currencies.FullCurrencies(ctx)
			if err == nil && (len(out.Currencies) != 1 || out.Currencies[0].Code != "AAVE") {
				t.Errorf("currencies = %+v", out)
			}
			return err
		}},
		{"Currencies.MerchantCoins", "currencies", "GET", "/v1/merchant/coins", "", func(t *testing.T, c *Client) error {
			_, err := c.Currencies.MerchantCoins(ctx)
			return err
		}},
		{"Estimates.Estimate", "estimate", "GET", "/v1/estimate?amount=3999.5&currency_from=usd&currency_to=btc", "", func(t *testing.T, c *Client) error {
			out, err := c.Estimates.Estimate(ctx, 3999.5, "usd", "btc")
			if err == nil && out.EstimatedAmount != 0.17061637 {
				t.Errorf("estimate = %+v", out)
			}
			return err
		}},
		{"Estimates.MinAmountEx", "min_amount", "GET", "/v1/min-amount?currency_from=eth&currency_to=trx&fiat_equivalent=usd&is_fixed_rate=true", "", func(t *testing.T, c *Client) error {
			out, err := c.Estimates.MinAmountEx(ctx, &models.MinAmountParams{CurrencyFrom: "eth", CurrencyTo: "trx", FiatEquivalent: "usd", IsFixedRate: true})
			if err == nil && (out.MinAmount != 0.0078999 || out.FiatEquivalent != 35.40626584) {
				t.Errorf("min amount = %+v", out)
			}
			return err
		}},
		{"Payouts.List", "payout_list", "GET", "/v1/payout?batch_id=5000145498&limit=20&status=finished", "", func(t *testing.T, c *Client) error {
			out, err := c.Payouts.List(ctx, &models.ListPayoutsParams{BatchID: "5000145498", Status: "finished", Limit: 20})
			if err == nil && (len(out) != 2 || out[0].Amount != 94.088939) {
				t.Errorf("payouts = %+v", out)
			}
			return err
		}},
		{"Payouts.ValidateAddress", "", "POST", "/v1/payout/validate-address", `{"address":"0x1","currency":"eth"}`, func(t *testing.T, c *Client) error {
			return c.Payouts.ValidateAddress(ctx, &models.ValidateAddressRequest{Address: "0x1", Currency: "eth"})
		}},
		{"Payouts.MinAmountForWithdrawal", "payout_min_amount", "GET", "/v1/payout-withdrawal/min-amount/trx", "", func(t *testing.T, c *Client) error {
			out, err := c.Payouts.MinAmountForWithdrawal(ctx, "trx")
			if err == nil && out != 0.00002496 {
				t.Errorf("min amount = %v", out)
			}
			return err
		}},
		{"Payouts.Fee", "payout_fee", "GET", "/v1/payout/fee?amount=500&currency=usdttrc20", "", func(t *testing.T, c *Client) error {
			out, err := c.Payouts.Fee(ctx, "usdttrc20", 500)
			if err == nil && out != 1.32765969 {
				t.Errorf("fee = %v", out)
			}
			return err
		}},
		{"Subscriptions.CreatePlan", "plan", "POST", "/v1/subscriptions/plans", `{"title":"test plan","interval_day":1,"amount":2,"currency":"usd"}`, func(t *testing.T, c *Client) error {
			out, err := c.Subscriptions.CreatePlan(ctx, &models.CreatePlanRequest{Title: "test plan", IntervalDay: 1, Amount: 2, Currency: "usd"})
			if err == nil && (out.ID != 76215585 || out.IntervalDay != 1) {
				t.Errorf("plan = %+v", out)
			}
			return err
		}},
		{"Subscriptions.UpdatePlan", "plan", "PATCH", "/v1/subscriptions/plans/76215585", `{"title":"test plan"}`, func(t *testing.T, c *Client) error {
			_, err := c.Subscriptions.UpdatePlan(ctx, "76215585", &models.UpdatePlanRequest{Title: "test plan"})
			return err
		}},
		{"Subscriptions.GetPlan", "plan", "GET", "/v1/subscriptions/plans/76215585", "", func(t *testing.T, c *Client) error {
			_, err := c.Subscriptions.GetPlan(ctx, "76215585")
			return err
		}},
		{"Subscriptions.ListPlans", "plans", "GET", "/v1/subscriptions/plans?limit=10&offset=3", "", func(t *testing.T, c *Client) error {
			out, err := c.Subscriptions.ListPlans(ctx, &models.ListPlansParams{Limit: 10, Offset: 3})
			if err == nil && len(out) != 2 {
				t.Errorf("plans = %+v", out)
			}
			return err
		}},
		{"Subscriptions.Create", "subscription", "POST", "/v1/subscriptions", `{"subscription_plan_id":76215585,"email":"test@example.com"}`, func(t *testing.T, c *Client) error {
			out, err := c.Subscriptions.Create(ctx, &models.CreateSubscriptionRequest{SubscriptionPlanID: 76215585, Email: "test@example.com"})
			if err == nil && (out.ID != 148427051 || out.Subscriber.Email != "test@example.com") {
				t.Errorf("subscription = %+v", out)
			}
			return err
		}},
		{"Subscriptions.Get", "subscription", "GET", "/v1/subscriptions/148427051", "", func(t *testing.T, c *Client) error {
			_, err := c.Subscriptions.Get(ctx, "148427051")
			return err
		}},
		{"Subscriptions.List", "subscriptions", "GET", "/v1/subscriptions?is_active=true&status=PAID", "", func(t *testing.T, c *Client) error {
			out, err := c.Subscriptions.List(ctx, &models.ListSubscriptionsParams{Status: "PAID", IsActive: &yes})
			if err == nil && (len(out) != 2 || out[0].Subscriber.SubPartnerID != 111394288) {
				t.Errorf("subscriptions = %+v", out)
			}
			return err
		}},
		{"Subscriptions.Delete", "subscription_delete", "DELETE", "/v1/subscriptions/148427051", "", func(t *testing.T, c *Client) error {
			return c.Subscriptions.Delete(ctx, "148427051")
		}},
		{"Conversions.List", "conversion_list", "GET", "/v1/conversion?limit=5&status=FINISHED", "", func(t *testing.T, c *Client) error {
			out, err := c.Conversions.List(ctx, &models.ListConversionsParams{Status: "FINISHED", Limit: 5})
			if err == nil && len(out) != 2 {
				t.Errorf("conversions = %+v", out)
			}
			return err
		}},
		{"Customers.Create", "customer_create", "POST", "/v1/sub-partner/balance", `{"name":"test1"}`, func(t *testing.T, c *Client) error {
			out, err := c.Customers.Create(ctx, &models.CreateCustomerRequest{Name: "test1"})
			if err == nil && out.ID != 1515573197 {
				t.Errorf("customer = %+v", out)
			}
			return err
		}},
		{"Customers.Balance", "customer_balance", "GET", "/v1/sub-partner/balance/111394288", "", func(t *testing.T, c *Client) error {
			out, err := c.Customers.Balance(ctx, "111394288")
			if err == nil && (out.SubPartnerID != 111394288 || out.Balances["usddtrc20"].Amount != 0.7) {
				t.Errorf("balance = %+v", out)
			}
			return err
		}},
		{"Customers.List", "customers", "GET", "/v1/sub-partner?limit=10&order=DESC", "", func(t *testing.T, c *Client) error {
			out, err := c.Customers.List(ctx, &models.ListCustomersParams{Limit: 10, Order: "DESC"})
			if err == nil && len(out) != 2 {
				t.Errorf("customers = %+v", out)
			}
			return err
		}},
		{"Customers.Payments", "customer_payments", "GET", "/v1/sub-partner/payments", "", func(t *testing.T, c *Client) error {
			_, err := c.Customers.Payments(ctx)
			return err
		}},
		{"Customers.Transfers", "transfers", "GET", "/v1/sub-partner/transfers?status=FINISHED", "", func(t *testing.T, c *Client) error {
			out, err := c.Customers.Transfers(ctx, &models.ListTransfersParams{Status: "FINISHED"})
			if err == nil && (len(out) != 2 || out[0].Amount != 1) {
				t.Errorf("transfers = %+v", out)
			}
			return err
		}},
		{"Customers.DepositWithPayment", "custody_payment", "POST", "/v1/sub-partner/payment", `{"currency":"trx","amount":0.3,"sub_partner_id":"1631380403"}`, func(t *testing.T, c *Client) error {
			out, err := c.Customers.DepositWithPayment(ctx, &models.CustodyDepositRequest{Currency: "trx", Amount: 0.3, SubPartnerID: "1631380403"})
			if err == nil && out.PaymentID != 5250038861 {
				t.Errorf("payment = %+v", out)
			}
			return err
		}},
		{"FiatPayouts.Providers", "fiat_providers", "GET", "/v1/fiat-payouts/providers", "", func(t *testing.T, c *Client) error {
			out, err := c.FiatPayouts.Providers(ctx)
			if err == nil && (len(out) != 1 || out[0].Code != "transfi") {
				t.Errorf("providers = %+v", out)
			}
			return err
		}},
		{"FiatPayouts.FiatCurrencies", "fiat_currencies", "GET", "/v1/fiat-payouts/fiat-currencies?provider=transfi", "", func(t *testing.T, c *Client) error {
			_, err := c.FiatPayouts.FiatCurrencies(ctx, "transfi")
			return err
		}},
		{"FiatPayouts.CryptoCurrencies", "fiat_crypto", "GET", "/v1/fiat-payouts/crypto-currencies?provider=transfi", "", func(t *testing.T, c *Client) error {
			_, err := c.FiatPayouts.CryptoCurrencies(ctx, "transfi", "")
			return err
		}},
		{"FiatPayouts.PaymentMethods", "fiat_methods", "GET", "/v1/fiat-payouts/payment-methods?currency=php&provider=transfi", "", func(t *testing.T, c *Client) error {
			out, err := c.FiatPayouts.PaymentMethods(ctx, "transfi", "php")
			if err == nil && (len(out) != 1 || len(out[0].Fields) != 1) {
				t.Errorf("methods = %+v", out)
			}
			return err
		}},
		{"FiatPayouts.Accounts", "fiat_accounts", "GET", "/v1/fiat-payouts/accounts", "", func(t *testing.T, c *Client) error {
			_, err := c.FiatPayouts.Accounts(ctx, "")
			return err
		}},
		{"FiatPayouts.Accounts by provider", "fiat_accounts", "GET", "/v1/fiat-payouts/accounts?provider=transfi", "", func(t *testing.T, c *Client) error {
			_, err := c.FiatPayouts.Accounts(ctx, "transfi")
			return err
		}},
		{"FiatPayouts.List", "fiat_payouts", "GET", "/v1/fiat-payouts?limit=10&status=FINISHED", "", func(t *testing.T, c *Client) error {
			out, err := c.FiatPayouts.List(ctx, &models.ListFiatPayoutsParams{Status: "FINISHED", Limit: 10})
			if err == nil && (len(out) != 1 || out[0].FiatAmount != "12.59") {
				t.Errorf("payouts = %+v", out)
			}
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.fixture != "" {
				body = fixture(t, tt.fixture)
			} else {
				body = []byte("OK")
			}
			c, rec := serve(t, http.StatusOK, body)
			if err := tt.call(t, c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.method != tt.method || rec.uri != tt.uri {
				t.Errorf("request = %s %s, want %s %s", rec.method, rec.uri, tt.method, tt.uri)
			}
			if tt.reqBody == "" {
				if rec.body != "" {
					t.Errorf("unexpected request body %q", rec.body)
				}
			} else {
				jsonEqual(t, rec.body, tt.reqBody)
			}
		})
	}
}

func TestRequestHeaders(t *testing.T) {
	c, rec := serve(t, http.StatusOK, fixture(t, "status"))
	if _, err := c.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := rec.header.Get("x-api-key"); got != "test-key" {
		t.Errorf("x-api-key = %q", got)
	}
	if got := rec.header.Get("Authorization"); got != "Bearer jwt" {
		t.Errorf("Authorization = %q", got)
	}
	if got := rec.header.Get("Content-Type"); got != "" {
		t.Errorf("GET sent Content-Type %q", got)
	}

	if _, err := c.Auth(context.Background(), "a", "b"); err != nil {
		t.Fatal(err)
	}
	if got := rec.header.Get("Content-Type"); got != "application/json" {
		t.Errorf("POST Content-Type = %q", got)
	}
}

func TestBaseURLTrailingSlash(t *testing.T) {
	var uri string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri = r.URL.Path
		_, _ = w.Write([]byte(`{"message":"OK"}`))
	}))
	defer srv.Close()
	c := NewClient(Config{APIKey: "k", BaseURL: srv.URL + "/v1/"})
	if _, err := c.Status(context.Background()); err != nil {
		t.Fatal(err)
	}
	if uri != "/v1/status" {
		t.Errorf("path = %q", uri)
	}
}

func TestAPIErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("400 carries code and message", func(t *testing.T) {
		c, _ := serve(t, http.StatusBadRequest, fixture(t, "error_400"))
		_, err := c.Subscriptions.Create(ctx, &models.CreateSubscriptionRequest{SubscriptionPlanID: 1})
		if !errors.Is(err, ErrBadRequest) {
			t.Fatalf("err = %v, want ErrBadRequest", err)
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) || !IsAPIError(err) {
			t.Fatalf("err is not an *APIError: %v", err)
		}
		if apiErr.StatusCode != 400 || apiErr.Code != "INVALID_REQUEST_PARAMS" || apiErr.Message == "" {
			t.Errorf("api error = %+v", apiErr)
		}
	})

	t.Run("404 matches generic and resource sentinels", func(t *testing.T) {
		c, _ := serve(t, http.StatusNotFound, fixture(t, "error_404"))
		_, payment := c.Payments.Get(ctx, 1)
		_, subscription := c.Subscriptions.Get(ctx, "1")
		deleted := c.Subscriptions.Delete(ctx, "1")
		_, invoice := c.Payments.CreateFromInvoice(ctx, &models.CreateInvoicePaymentRequest{IID: 1})

		for _, tc := range []struct {
			name      string
			err, want error
		}{
			{"payment", payment, ErrPaymentNotFound},
			{"subscription", subscription, ErrSubscriptionNotFound},
			{"delete", deleted, ErrSubscriptionNotFound},
			{"invoice", invoice, ErrInvoiceNotFound},
		} {
			if !errors.Is(tc.err, tc.want) || !errors.Is(tc.err, ErrNotFound) || !IsAPIError(tc.err) {
				t.Errorf("%s: err = %v", tc.name, tc.err)
			}
		}
	})

	t.Run("401 is unauthorized and invalid key", func(t *testing.T) {
		c, _ := serve(t, http.StatusUnauthorized, []byte(`{"message":"Invalid api key"}`))
		_, err := c.Balance(ctx)
		if !errors.Is(err, ErrUnauthorized) || !errors.Is(err, ErrInvalidAPIKey) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("plain text body", func(t *testing.T) {
		c, _ := serve(t, http.StatusBadGateway, []byte("bad gateway"))
		_, err := c.Status(ctx)
		var apiErr *APIError
		if !errors.Is(err, ErrServerError) || !errors.As(err, &apiErr) || apiErr.Message != "bad gateway" {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("nil MinAmountEx params", func(t *testing.T) {
		c, _ := serve(t, http.StatusOK, nil)
		if _, err := c.Estimates.MinAmountEx(ctx, nil); err == nil {
			t.Error("want an error for nil params")
		}
	})
}

func TestRetries(t *testing.T) {
	ctx := context.Background()
	retry := RetryConfig{MaxRetries: 2, InitialBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond}

	newServer := func(t *testing.T, fail int, status int) (*Client, *atomic.Int32) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if int(calls.Add(1)) <= fail {
				w.WriteHeader(status)
				return
			}
			_, _ = w.Write(fixture(t, "payment_create"))
		}))
		t.Cleanup(srv.Close)
		return NewClient(Config{APIKey: "k", BaseURL: srv.URL, Retry: retry}), &calls
	}

	t.Run("GET retries 5xx", func(t *testing.T) {
		c, calls := newServer(t, 2, http.StatusServiceUnavailable)
		if _, err := c.Payments.Get(ctx, 1); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 3 {
			t.Errorf("calls = %d, want 3", calls.Load())
		}
	})

	t.Run("POST does not retry 5xx", func(t *testing.T) {
		c, calls := newServer(t, 1, http.StatusInternalServerError)
		_, err := c.Payments.Create(ctx, &models.CreatePaymentRequest{})
		if !errors.Is(err, ErrServerError) {
			t.Fatalf("err = %v", err)
		}
		if calls.Load() != 1 {
			t.Errorf("calls = %d, want 1", calls.Load())
		}
	})

	t.Run("POST retries 429", func(t *testing.T) {
		c, calls := newServer(t, 1, http.StatusTooManyRequests)
		if _, err := c.Payments.Create(ctx, &models.CreatePaymentRequest{}); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 2 {
			t.Errorf("calls = %d, want 2", calls.Load())
		}
	})

	t.Run("gives up after MaxRetries", func(t *testing.T) {
		c, calls := newServer(t, 10, http.StatusTooManyRequests)
		_, err := c.Status(ctx)
		if !errors.Is(err, ErrRateLimited) {
			t.Fatalf("err = %v", err)
		}
		if calls.Load() != 3 {
			t.Errorf("calls = %d, want 3", calls.Load())
		}
	})

	t.Run("no retries by default", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()
		c := NewClient(Config{APIKey: "k", BaseURL: srv.URL})
		if _, err := c.Status(ctx); err == nil {
			t.Fatal("want an error")
		}
		if calls.Load() != 1 {
			t.Errorf("calls = %d, want 1", calls.Load())
		}
	})

	t.Run("context cancels the wait", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()
		c := NewClient(Config{APIKey: "k", BaseURL: srv.URL, Retry: RetryConfig{MaxRetries: 5, InitialBackoff: time.Hour}})
		ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancel()
		if _, err := c.Status(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("err = %v, want deadline exceeded", err)
		}
	})
}

func TestNewClientRequiresAPIKey(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewClient with no API key did not panic")
		}
	}()
	NewClient(Config{})
}
