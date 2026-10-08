//go:build live

// Live tests against a real NOWPayments account. They are excluded from
// normal builds; run them with:
//
//	NOWPAYMENTS_API_KEY=... go test -tags live -run Live -v -count=1 .
//
// Environment:
//
//	NOWPAYMENTS_API_KEY      required
//	NOWPAYMENTS_BASE_URL     defaults to the sandbox
//	NOWPAYMENTS_EMAIL        with NOWPAYMENTS_PASSWORD, enables the JWT endpoints
//	NOWPAYMENTS_PASSWORD
//	NOWPAYMENTS_LIVE_CREATE  on a production account, set to 1 to also create payments, invoices, plans and customers
//	NOWPAYMENTS_LIVE_STRICT  set to 1 to fail instead of skip when the API refuses a call or a step cannot run
//	NOWPAYMENTS_LIVE_DUMP    folder for raw responses, default testdata/live
//
// Every raw response is saved to the dump folder so the fixtures in
// testdata can be checked against what the API really returns. The JWT
// from /auth is redacted; the API key is never written.
package nowpay

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/muxover/nowpay-go/models"
)

type dumpTransport struct {
	dir  string
	next http.RoundTripper
	n    atomic.Int32
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func (d *dumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := d.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	saved := body
	if strings.HasSuffix(req.URL.Path, "/auth") && resp.StatusCode < 300 {
		saved = []byte(`{"token":"<redacted>"}`)
	}
	name := fmt.Sprintf("%02d_%s_%s_%d.json", d.n.Add(1), req.Method,
		strings.Trim(unsafeChars.ReplaceAllString(req.URL.Path, "_"), "_"), resp.StatusCode)
	record := fmt.Sprintf("%s %s\n%d\n%s\n", req.Method, req.URL.RequestURI(), resp.StatusCode, saved)
	if err := os.WriteFile(filepath.Join(d.dir, name), []byte(record), 0o644); err != nil {
		return nil, err
	}
	return resp, nil
}

type liveEnv struct {
	client  *Client
	jwt     *Client
	sandbox bool
	create  bool
}

// liveStrict turns every skip into a failure, so a passing run proves that
// every method was called and answered.
var liveStrict bool

func setupLive(t *testing.T) *liveEnv {
	t.Helper()
	liveStrict = os.Getenv("NOWPAYMENTS_LIVE_STRICT") == "1"
	key := os.Getenv("NOWPAYMENTS_API_KEY")
	if key == "" {
		skip(t, "NOWPAYMENTS_API_KEY is not set")
	}
	base := os.Getenv("NOWPAYMENTS_BASE_URL")
	if base == "" {
		base = SandboxBaseURL
	}
	dir := os.Getenv("NOWPAYMENTS_LIVE_DUMP")
	if dir == "" {
		dir = filepath.Join("testdata", "live")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	httpClient := &http.Client{Timeout: 60 * time.Second, Transport: &dumpTransport{dir: dir, next: http.DefaultTransport}}
	env := &liveEnv{
		sandbox: strings.Contains(base, "sandbox"),
		client:  NewClient(Config{APIKey: key, BaseURL: base, HTTPClient: httpClient}),
	}
	env.create = env.sandbox || os.Getenv("NOWPAYMENTS_LIVE_CREATE") == "1"
	t.Logf("base URL %s (sandbox: %v), raw responses in %s", base, env.sandbox, dir)

	if email, pw := os.Getenv("NOWPAYMENTS_EMAIL"), os.Getenv("NOWPAYMENTS_PASSWORD"); email != "" && pw != "" {
		auth, err := env.client.Auth(context.Background(), email, pw)
		if err != nil {
			skip(t, fmt.Sprintf("Auth failed, JWT endpoints will be skipped: %v", err))
		} else {
			env.jwt = NewClient(Config{APIKey: key, Token: auth.Token, BaseURL: base, HTTPClient: httpClient})
		}
	}
	return env
}

// check fails the test on SDK errors (bad requests we built, responses we
// could not decode) and skips on API errors the account may legitimately hit,
// such as a feature not enabled in the sandbox. The raw dump shows which.
func check(t *testing.T, err error) bool {
	t.Helper()
	if err == nil {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		skip(t, fmt.Sprintf("API answered %d %s: %s", apiErr.StatusCode, apiErr.Code, apiErr.Message))
	}
	t.Fatalf("SDK error: %v", err)
	return false
}

// skip skips the test, or fails it in strict mode.
func skip(t *testing.T, reason string) {
	t.Helper()
	if liveStrict {
		t.Fatal(reason)
	}
	t.Skip(reason)
}

func needJWT(t *testing.T, env *liveEnv) *Client {
	t.Helper()
	if env.jwt == nil {
		skip(t, "set NOWPAYMENTS_EMAIL and NOWPAYMENTS_PASSWORD for this endpoint")
	}
	return env.jwt
}

// needCreate skips calls that leave records behind (a pending payment, a
// plan, a custody customer) when running against a production account.
func needCreate(t *testing.T, env *liveEnv) {
	t.Helper()
	if !env.create {
		skip(t, "creates a record on a production account; set NOWPAYMENTS_LIVE_CREATE=1 to run it")
	}
}

func TestLive(t *testing.T) {
	env := setupLive(t)
	ctx := context.Background()
	c := env.client
	tag := fmt.Sprintf("nowpay-go-live-%d", time.Now().Unix())

	t.Run("Status", func(t *testing.T) {
		out, err := c.Status(ctx)
		if check(t, err) && out.Message != "OK" {
			t.Errorf("status = %+v", out)
		}
	})

	var coin string
	t.Run("Currencies", func(t *testing.T) {
		list, err := c.Currencies.Supported(ctx)
		if !check(t, err) {
			return
		}
		if len(list) == 0 {
			t.Fatal("no currencies")
		}
		for _, want := range []string{"btc", "trx", "eth", "ltc"} {
			for _, got := range list {
				if got == want && coin == "" {
					coin = want
				}
			}
		}
		if coin == "" {
			coin = list[0]
		}
		t.Logf("%d currencies, using %s", len(list), coin)
		fixed, err := c.Currencies.SupportedWithFixedRate(ctx, true)
		if check(t, err) && (len(fixed) == 0 || fixed[0] == "") {
			t.Errorf("fixed-rate currencies = %v", fixed)
		}
	})
	if coin == "" {
		coin = "btc"
	}

	t.Run("FullCurrencies", func(t *testing.T) {
		out, err := c.Currencies.FullCurrencies(ctx)
		if check(t, err) && len(out.Currencies) == 0 {
			t.Error("empty list")
		}
	})
	t.Run("MerchantCoins", func(t *testing.T) {
		out, err := c.Currencies.MerchantCoins(ctx)
		if check(t, err) && len(out) == 0 {
			t.Error("no merchant coins decoded")
		}
	})
	t.Run("Estimate", func(t *testing.T) {
		out, err := c.Estimates.Estimate(ctx, 100, "usd", coin)
		if check(t, err) && out.EstimatedAmount <= 0 {
			t.Errorf("estimate = %+v", out)
		}
	})
	t.Run("MinAmountEx", func(t *testing.T) {
		out, err := c.Estimates.MinAmountEx(ctx, &models.MinAmountParams{CurrencyFrom: coin, CurrencyTo: coin, FiatEquivalent: "usd"})
		if check(t, err) && out.MinAmount <= 0 {
			t.Errorf("min amount = %+v", out)
		}
	})

	var paymentID int64
	t.Run("Payments.Create", func(t *testing.T) {
		needCreate(t, env)
		out, err := c.Payments.Create(ctx, &models.CreatePaymentRequest{
			PriceAmount: 100, PriceCurrency: "usd", PayCurrency: coin,
			OrderID: tag, OrderDescription: "nowpay-go live test",
		})
		if !check(t, err) {
			return
		}
		if out.PaymentID == 0 || out.PayAddress == "" || out.PaymentStatus == "" {
			t.Errorf("payment = %+v", out)
		}
		paymentID = out.PaymentID.Int64()
	})
	t.Run("Payments.Get", func(t *testing.T) {
		if paymentID == 0 && env.jwt != nil {
			if list, err := env.jwt.Payments.List(ctx, &models.ListPaymentsParams{Limit: 1}); err == nil && len(list) > 0 {
				paymentID = list[0].PaymentID.Int64()
			}
		}
		if paymentID == 0 {
			skip(t, "no payment to read")
		}
		out, err := c.Payments.Get(ctx, paymentID)
		if check(t, err) && out.PaymentID.Int64() != paymentID {
			t.Errorf("payment = %+v", out)
		}
	})
	t.Run("Payments.UpdateMerchantEstimate", func(t *testing.T) {
		if paymentID == 0 {
			skip(t, "no payment created")
		}
		_, err := c.Payments.UpdateMerchantEstimate(ctx, paymentID)
		check(t, err)
	})
	t.Run("Payments.Get unknown", func(t *testing.T) {
		_, err := c.Payments.Get(ctx, 1)
		if !errors.Is(err, ErrPaymentNotFound) || !IsAPIError(err) {
			t.Fatalf("want ErrPaymentNotFound, got %v", err)
		}
	})

	var invoiceID int64
	t.Run("Invoices.Create", func(t *testing.T) {
		needCreate(t, env)
		out, err := c.Invoices.Create(ctx, &models.CreateInvoiceRequest{
			PriceAmount: 100, PriceCurrency: "usd", OrderID: tag + "-inv", OrderDescription: "nowpay-go live test",
		})
		if !check(t, err) {
			return
		}
		if out.ID == 0 || out.InvoiceURL == "" {
			t.Errorf("invoice = %+v", out)
		}
		invoiceID = out.ID
	})
	t.Run("Payments.CreateFromInvoice", func(t *testing.T) {
		if invoiceID == 0 {
			skip(t, "no invoice created")
		}
		out, err := c.Payments.CreateFromInvoice(ctx, &models.CreateInvoicePaymentRequest{IID: invoiceID, PayCurrency: coin, OrderDescription: "nowpay-go live test"})
		if check(t, err) && out.PaymentID == 0 {
			t.Errorf("payment = %+v", out)
		}
	})
	t.Run("Payments.List", func(t *testing.T) {
		out, err := needJWT(t, env).Payments.List(ctx, &models.ListPaymentsParams{Limit: 5})
		if check(t, err) {
			t.Logf("%d payments", len(out))
			for _, p := range out {
				if p.PaymentID == 0 || p.PaymentStatus == "" {
					t.Errorf("incomplete payment: %+v", p)
				}
			}
		}
	})
	t.Run("Balance", func(t *testing.T) {
		out, err := needJWT(t, env).Balance(ctx)
		if check(t, err) {
			t.Logf("%d currencies with a balance", len(out.Currencies))
		}
	})

	// Payouts
	t.Run("Payouts.ValidateAddress invalid", func(t *testing.T) {
		err := needJWT(t, env).Payouts.ValidateAddress(ctx, &models.ValidateAddressRequest{Address: "not-an-address", Currency: coin})
		if err == nil {
			t.Fatal("an invalid address was accepted")
		}
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Fatalf("want an APIError, got %v", err)
		}
		t.Logf("-> %d %s %q", apiErr.StatusCode, apiErr.Code, apiErr.Message)
	})
	t.Run("Payouts.MinAmountForWithdrawal", func(t *testing.T) {
		out, err := needJWT(t, env).Payouts.MinAmountForWithdrawal(ctx, coin)
		if check(t, err) && out <= 0 {
			t.Errorf("min amount = %v", out)
		}
	})
	t.Run("Payouts.Fee", func(t *testing.T) {
		_, err := needJWT(t, env).Payouts.Fee(ctx, coin, 1)
		check(t, err)
	})
	t.Run("Payouts.List", func(t *testing.T) {
		out, err := needJWT(t, env).Payouts.List(ctx, &models.ListPayoutsParams{Limit: 5})
		if check(t, err) {
			t.Logf("%d payouts", len(out))
		}
	})

	// Recurring payments
	var planID string
	t.Run("Subscriptions.CreatePlan", func(t *testing.T) {
		jwt := needJWT(t, env)
		// Reuse the plan from an earlier run instead of piling up new ones.
		if plans, err := jwt.Subscriptions.ListPlans(ctx, &models.ListPlansParams{Limit: 100}); err == nil {
			for _, p := range plans {
				if strings.HasPrefix(p.Title, "nowpay-go-live") {
					planID = id(p.ID)
					t.Logf("reusing plan %s from an earlier run", planID)
					return
				}
			}
		}
		needCreate(t, env)
		out, err := jwt.Subscriptions.CreatePlan(ctx, &models.CreatePlanRequest{Title: tag, IntervalDay: 30, Amount: 10, Currency: "usd"})
		if check(t, err) {
			planID = id(out.ID)
		}
	})
	t.Run("Subscriptions.Plans", func(t *testing.T) {
		if planID == "" {
			skip(t, "no plan created")
		}
		_, err := env.jwt.Subscriptions.GetPlan(ctx, planID)
		check(t, err)
		_, err = env.jwt.Subscriptions.UpdatePlan(ctx, planID, &models.UpdatePlanRequest{Title: tag + "-updated"})
		check(t, err)
		_, err = env.jwt.Subscriptions.ListPlans(ctx, &models.ListPlansParams{Limit: 5})
		check(t, err)
	})
	t.Run("Subscriptions.Email", func(t *testing.T) {
		if planID == "" {
			skip(t, "no plan created")
		}
		sub, err := env.jwt.Subscriptions.Create(ctx, &models.CreateSubscriptionRequest{SubscriptionPlanID: mustInt(t, planID), Email: liveEmail})
		if !check(t, err) {
			return
		}
		got, err := env.jwt.Subscriptions.Get(ctx, id(sub.ID))
		if check(t, err) && got.ID != sub.ID {
			t.Errorf("got %+v", got)
		}
	})
	t.Run("Subscriptions.ListAndCleanUp", func(t *testing.T) {
		subs, err := needJWT(t, env).Subscriptions.List(ctx, &models.ListSubscriptionsParams{Limit: 100})
		if !check(t, err) {
			return
		}
		deleted := 0
		for _, s := range subs {
			if s.Subscriber.Email == liveEmail {
				if check(t, env.jwt.Subscriptions.Delete(ctx, id(s.ID))) {
					deleted++
				}
			}
		}
		t.Logf("%d subscriptions, deleted %d test ones", len(subs), deleted)
	})

	// Custody
	var customerID string
	t.Run("Customers.Create", func(t *testing.T) {
		jwt := needJWT(t, env)
		// Custody customers cannot be deleted, so reuse one from an earlier run.
		if list, err := jwt.Customers.List(ctx, &models.ListCustomersParams{Limit: 100}); err == nil {
			for _, c := range list {
				if strings.HasPrefix(c.Name, "nowpay-go-live") {
					customerID = id(c.ID)
					t.Logf("reusing customer %s from an earlier run", customerID)
					return
				}
			}
		}
		needCreate(t, env)
		out, err := jwt.Customers.Create(ctx, &models.CreateCustomerRequest{Name: tag})
		if check(t, err) {
			customerID = id(out.ID)
		}
	})
	t.Run("Customers.Balance", func(t *testing.T) {
		if customerID == "" {
			skip(t, "no customer created")
		}
		_, err := env.jwt.Customers.Balance(ctx, customerID)
		check(t, err)
	})
	t.Run("Customers.List", func(t *testing.T) {
		out, err := needJWT(t, env).Customers.List(ctx, &models.ListCustomersParams{Limit: 5})
		if check(t, err) {
			t.Logf("%d customers", len(out))
		}
	})
	t.Run("Customers.Transfers", func(t *testing.T) {
		_, err := needJWT(t, env).Customers.Transfers(ctx, &models.ListTransfersParams{Limit: 5})
		check(t, err)
	})
	t.Run("Customers.Payments", func(t *testing.T) {
		out, err := needJWT(t, env).Customers.Payments(ctx)
		if check(t, err) {
			for _, p := range out {
				if p.PaymentID == 0 || p.PaymentStatus == "" {
					t.Errorf("incomplete custody payment: %+v", p)
				}
			}
		}
	})
	t.Run("Customers.DepositWithPayment", func(t *testing.T) {
		if customerID == "" {
			skip(t, "no customer created")
		}
		out, err := env.jwt.Customers.DepositWithPayment(ctx, &models.CustodyDepositRequest{Currency: coin, Amount: 1, SubPartnerID: customerID})
		if check(t, err) && out.PaymentID == 0 {
			t.Errorf("payment = %+v", out)
		}
	})

	// Conversions
	t.Run("Conversions.List", func(t *testing.T) {
		_, err := needJWT(t, env).Conversions.List(ctx, &models.ListConversionsParams{Limit: 5})
		check(t, err)
	})

	// Fiat payouts
	provider := ""
	t.Run("FiatPayouts.Providers", func(t *testing.T) {
		out, err := needJWT(t, env).FiatPayouts.Providers(ctx)
		if check(t, err) && len(out) > 0 {
			provider = out[0].Code
		}
	})
	fiat := func(name string, fn func(c *Client) error) {
		t.Run("FiatPayouts."+name, func(t *testing.T) {
			jwt := needJWT(t, env)
			if provider == "" {
				skip(t, "no fiat provider")
			}
			check(t, fn(jwt))
		})
	}
	var fiatCurrency string
	fiat("FiatCurrencies", func(c *Client) error {
		out, err := c.FiatPayouts.FiatCurrencies(ctx, provider)
		if err == nil && len(out) > 0 {
			fiatCurrency = out[0].CurrencyCode
		}
		return err
	})
	fiat("CryptoCurrencies", func(c *Client) error {
		_, err := c.FiatPayouts.CryptoCurrencies(ctx, provider, fiatCurrency)
		return err
	})
	fiat("PaymentMethods", func(c *Client) error {
		_, err := c.FiatPayouts.PaymentMethods(ctx, provider, fiatCurrency)
		return err
	})
	fiat("Accounts", func(c *Client) error {
		_, err := c.FiatPayouts.Accounts(ctx, provider)
		return err
	})
	fiat("List", func(c *Client) error {
		_, err := c.FiatPayouts.List(ctx, &models.ListFiatPayoutsParams{Limit: 5})
		return err
	})
}

const liveEmail = "nowpay-go-live@example.com"

func id(n models.FlexInt) string { return strconv.FormatInt(n.Int64(), 10) }

func mustInt(t *testing.T, s string) int64 {
	t.Helper()
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		t.Fatalf("id %q is not numeric", s)
	}
	return v
}
