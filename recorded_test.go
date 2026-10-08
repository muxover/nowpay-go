package nowpay

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/muxover/nowpay-go/models"
)

// TestRecordedResponses runs the services against real responses captured
// from the production API (testdata/recorded). Where these differ from the
// documented examples, the documented shape is covered in TestEndpoints.
func TestRecordedResponses(t *testing.T) {
	ctx := context.Background()
	rec := func(t *testing.T, name string, status int) *Client {
		c, _ := serve(t, status, fixture(t, "recorded/"+name))
		return c
	}

	t.Run("fixed-rate currencies are objects", func(t *testing.T) {
		out, err := rec(t, "currencies_fixed_rate", 200).Currencies.SupportedWithFixedRate(ctx, true)
		if err != nil || len(out) != 3 || out[0] != "aave" {
			t.Errorf("got %v, %v", out, err)
		}
	})
	t.Run("merchant coins use selectedCurrencies", func(t *testing.T) {
		out, err := rec(t, "merchant_coins", 200).Currencies.MerchantCoins(ctx)
		if err != nil || len(out) != 3 || out[0] != "BTC" {
			t.Errorf("got %v, %v", out, err)
		}
	})
	t.Run("full currencies send network_precision as a string", func(t *testing.T) {
		out, err := rec(t, "full_currencies", 200).Currencies.FullCurrencies(ctx)
		if err != nil || len(out.Currencies) != 2 {
			t.Fatalf("got %+v, %v", out, err)
		}
		c := out.Currencies[0]
		if c.NetworkPrecision == nil || *c.NetworkPrecision != 18 || !c.AvailableForPayment || c.Precision != 8 {
			t.Errorf("currency = %+v", c)
		}
	})
	t.Run("estimate sends estimated_amount as a string", func(t *testing.T) {
		out, err := rec(t, "estimate", 200).Estimates.Estimate(ctx, 100, "usd", "btc")
		if err != nil || out.EstimatedAmount != 0.00118682 || out.AmountFrom != 100 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("min amount", func(t *testing.T) {
		out, err := rec(t, "min_amount", 200).Estimates.MinAmount(ctx, "btc", "btc")
		if err != nil || out != 0.0002625 {
			t.Errorf("got %v, %v", out, err)
		}
	})
	t.Run("payment create", func(t *testing.T) {
		out, err := rec(t, "payment_create", 201).Payments.Create(ctx, &models.CreatePaymentRequest{PriceAmount: 100, PriceCurrency: "usd", PayCurrency: "btc"})
		if err != nil || out.PaymentID != 6291776732 || out.PurchaseID != 4933037607 || out.PayAmount != 0.00118666 || out.Product != "api" {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("payment get", func(t *testing.T) {
		out, err := rec(t, "payment_get", 200).Payments.Get(ctx, 6291776732)
		if err != nil || out.PaymentStatus != "waiting" || out.OutcomeCurrency != "usdttrc20" {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("merchant estimate", func(t *testing.T) {
		out, err := rec(t, "update_estimate", 200).Payments.UpdateMerchantEstimate(ctx, 6291776732)
		if err != nil || out.ID != 6291776732 || out.PayAmount != 0.00118666 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("invoice create", func(t *testing.T) {
		out, err := rec(t, "invoice_create", 200).Invoices.Create(ctx, &models.CreateInvoiceRequest{PriceAmount: 100, PriceCurrency: "usd"})
		if err != nil || out.ID != 4595166522 || out.TokenID != "5171235646" || out.PriceAmount != 100 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("payment from invoice", func(t *testing.T) {
		out, err := rec(t, "invoice_payment", 201).Payments.CreateFromInvoice(ctx, &models.CreateInvoicePaymentRequest{IID: 4595166522, PayCurrency: "btc"})
		if err != nil || out.PaymentID != 5561513183 || out.Product != "invoice" {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("subscription create returns an array", func(t *testing.T) {
		out, err := rec(t, "subscription_create", 200).Subscriptions.Create(ctx, &models.CreateSubscriptionRequest{SubscriptionPlanID: 1143462121, Email: "a@example.com"})
		if err != nil || out.ID != 943421599 || out.Status != "WAITING_PAY" {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("plan", func(t *testing.T) {
		out, err := rec(t, "plan", 200).Subscriptions.GetPlan(ctx, "1143462121")
		if err != nil || out.IntervalDay != 30 || out.Amount != 10 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("custody payment", func(t *testing.T) {
		out, err := rec(t, "custody_payment", 200).Customers.DepositWithPayment(ctx, &models.CustodyDepositRequest{Currency: "btc", Amount: 1, SubPartnerID: "356992164"})
		if err != nil || out.PaymentID != 5373092876 || out.InvoiceID != 6414848611 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("custody payments use status and text prices", func(t *testing.T) {
		out, err := rec(t, "customer_payments", 200).Customers.Payments(ctx)
		if err != nil || len(out) != 1 {
			t.Fatalf("got %+v, %v", out, err)
		}
		p := out[0]
		if p.PaymentStatus != "waiting" || p.PayAmount != 1 || p.OriginalPrice != "1 BTC" || p.SubPartnerID != 356992164 || len(p.ExchangeData) == 0 {
			t.Errorf("payment = %+v", p)
		}
	})
	t.Run("payout fee", func(t *testing.T) {
		out, err := rec(t, "payout_fee", 200).Payouts.Fee(ctx, "btc", 1)
		if err != nil || out != 0.0000432 {
			t.Errorf("got %v, %v", out, err)
		}
	})
	t.Run("fiat providers", func(t *testing.T) {
		out, err := rec(t, "fiat_providers", 200).FiatPayouts.Providers(ctx)
		if err != nil || len(out) != 2 || out[0].Code != "guardarian" {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("IP not whitelisted", func(t *testing.T) {
		_, err := rec(t, "error_403_ip", http.StatusForbidden).Balance(ctx)
		var apiErr *APIError
		if !errors.Is(err, ErrForbidden) || !errors.As(err, &apiErr) || apiErr.Code != "ENDPOINT_NOT_ALLOWED" {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("unknown payment", func(t *testing.T) {
		_, err := rec(t, "error_404_payment", http.StatusNotFound).Payments.Get(ctx, 1)
		var apiErr *APIError
		if !errors.Is(err, ErrPaymentNotFound) || !errors.As(err, &apiErr) || apiErr.Message != "Payment not found" {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("server error keeps code and message", func(t *testing.T) {
		_, err := rec(t, "error_500_order_by", http.StatusInternalServerError).Payments.List(ctx, nil)
		var apiErr *APIError
		if !errors.Is(err, ErrServerError) || !errors.As(err, &apiErr) || apiErr.Code != "INTERNAL_ERROR" || apiErr.Message != "Invalid order_by value" {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("account balance", func(t *testing.T) {
		out, err := rec(t, "balance", 200).Balance(ctx)
		if err != nil || out.Currencies["usdttrc20"].Amount != 1.5 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
	t.Run("withdrawal minimum is under result", func(t *testing.T) {
		out, err := rec(t, "payout_min_amount", 200).Payouts.MinAmountForWithdrawal(ctx, "btc")
		if err != nil || out != 0.000016 {
			t.Errorf("got %v, %v", out, err)
		}
	})
	t.Run("empty customer balance is an array", func(t *testing.T) {
		out, err := rec(t, "customer_balance_empty", 200).Customers.Balance(ctx, "111394288")
		if err != nil || out.SubPartnerID != 111394288 || out.Balances == nil || len(out.Balances) != 0 {
			t.Errorf("got %+v, %v", out, err)
		}
	})
}
