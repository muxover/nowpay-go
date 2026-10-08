# NowPay Go

<div align="center">

[![CI](https://github.com/muxover/nowpay-go/actions/workflows/ci.yml/badge.svg)](https://github.com/muxover/nowpay-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/muxover/nowpay-go.svg)](https://pkg.go.dev/github.com/muxover/nowpay-go)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/muxover/nowpay-go)](https://github.com/muxover/nowpay-go/releases/latest)

**Go SDK for NOWPayments: payments, custody, and webhooks.**

</div>

---

NowPay Go is a Go client for the [NOWPayments](https://nowpayments.io) crypto payment gateway. It covers payments, invoices, currencies, estimates, recurring payments, custody customers, and the payout, conversion, and fiat-payout lookups, and it verifies and parses IPN callbacks. Every method has been run against the production API. It has no dependencies outside the standard library.

---

## Features

- Typed requests and responses for every method, each one run against the production NOWPayments API (see [API status](#api-status)).
- IPN signature checks that match how NOWPayments signs callbacks, including payloads with `&`, `<`, `>`, and non-ASCII text.
- Typed payment events, with the event type derived from the payment status.
- Errors you can match with `errors.Is`, and an `APIError` that carries the status, code, and message.
- Optional retries that never repeat a `POST` after a server error, so a payment is not created twice.
- Handles the places where the live API differs from its docs: IDs and amounts sent as numbers or strings, empty balances sent as `[]`, and responses wrapped in different shapes.
- `context.Context` on every call, a configurable `*http.Client`, and a sandbox base URL.

---

## Installation

```bash
go get github.com/muxover/nowpay-go
```

Requires Go 1.21 or later.

---

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/muxover/nowpay-go"
	"github.com/muxover/nowpay-go/models"
)

func main() {
	client := nowpay.NewClient(nowpay.Config{APIKey: os.Getenv("NOWPAYMENTS_API_KEY")})

	payment, err := client.Payments.Create(context.Background(), &models.CreatePaymentRequest{
		PriceAmount:    10,
		PriceCurrency:  "usd",
		PayCurrency:    "btc",
		OrderID:        "order-001",
		IPNCallbackURL: "https://example.com/webhook",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(payment.PaymentID, payment.PayAddress, payment.PayAmount)
}
```

---

## API overview

| Service | Methods |
|---------|---------|
| **Client** | `Status`, `Auth` (JWT from email and password), `Balance` |
| **Payments** | `Create`, `CreateFromInvoice`, `Get`, `List`, `UpdateMerchantEstimate` |
| **Invoices** | `Create` |
| **Currencies** | `Supported`, `SupportedWithFixedRate`, `FullCurrencies`, `MerchantCoins` |
| **Estimates** | `Estimate`, `EstimateByPrice`, `MinAmount`, `MinAmountEx` |
| **Payouts** | `List`, `ValidateAddress`, `MinAmountForWithdrawal`, `Fee` |
| **Subscriptions** | `CreatePlan`, `UpdatePlan`, `GetPlan`, `ListPlans`, `Create`, `Get`, `List`, `Delete` |
| **Conversions** | `List` |
| **Customers** (custody) | `Create`, `Balance`, `List`, `Payments`, `Transfers`, `DepositWithPayment` |
| **FiatPayouts** | `Providers`, `FiatCurrencies`, `CryptoCurrencies`, `PaymentMethods`, `Accounts`, `List` |

Payouts, conversions, custody, fiat payouts, recurring payments, and on most accounts `Payments.List` need a JWT. Call `Auth(ctx, email, password)` and pass the token as `Config.Token`. The token expires after about five minutes, so fetch a new one for each batch of calls.

If your account restricts API access to whitelisted IPs, `Balance`, `Payouts.MinAmountForWithdrawal`, and `Customers.Balance` answer `403 ENDPOINT_NOT_ALLOWED` from any other IP. Add the server's IP in the NOWPayments dashboard.

---

## API status

Every method above was run against the production NOWPayments API with `live_test.go` in strict mode, where any refused or skipped call fails the run. The real responses are kept in `testdata/recorded/` and tested on every change.

These NOWPayments endpoints are not in the SDK, because they move funds and cannot be tested without doing so: creating, verifying, and cancelling payouts; custody transfers, deposits, and write-offs; creating conversions; and fiat-payout accounts and requests. They will be added once each has been verified live.

Webhook signatures are checked against callbacks signed with Node.js the way NOWPayments signs them.

---

## Project Layout

```text
nowpay-go/
├── doc.go                 # Package overview and quick start
├── client.go              # Client, Config, and every service
├── internal/
│   ├── http.go            # Transport: auth headers, JSON, retries
│   └── errors.go          # Sentinel errors and APIError
├── models/                # Request and response types, FlexInt / FlexFloat
├── webhook/
│   ├── verify.go          # IPN HMAC-SHA512 signature checks
│   └── events.go          # Payment callback parsing
├── client_test.go         # Every service against the documented responses
├── recorded_test.go       # Every service against recorded live responses
├── live_test.go           # Opt-in tests against a real account
├── testdata/              # Documented and recorded API responses
├── examples/              # Create a payment; verify and parse IPN callbacks
├── release-notes/         # One file per release
└── docs/PROJECT.md        # Maintainer notes
```

---

## Configuration

| Field | Description |
|-------|-------------|
| `APIKey` | Required. Your NOWPayments API key. `NewClient` panics without it. |
| `Token` | Optional. JWT from `Auth()` for endpoints that need one. |
| `BaseURL` | Optional. Defaults to `nowpay.DefaultBaseURL`. Use `nowpay.SandboxBaseURL` with a sandbox key. |
| `Timeout` | Optional. HTTP timeout, 30s by default. Ignored when `HTTPClient` is set. |
| `HTTPClient` | Optional. Your own `*http.Client`. |
| `Retry` | Optional. `MaxRetries`, `InitialBackoff`, `MaxBackoff`. Zero means no retries. |

With retries on, a `429` is retried for every request and honors `Retry-After`. A `5xx` is retried only for `GET`, `PUT`, and `DELETE`. A `POST` that failed with a server error may still have gone through, so it is returned to you rather than sent again.

```go
client := nowpay.NewClient(nowpay.Config{
	APIKey: apiKey,
	Retry: nowpay.RetryConfig{
		MaxRetries:     3,
		InitialBackoff: time.Second,
		MaxBackoff:     30 * time.Second,
	},
})
```

---

## Estimates and fiat

```go
est, err := client.Estimates.Estimate(ctx, 25, "usd", "btc")
// est.EstimatedAmount is the amount in BTC

min, err := client.Estimates.MinAmountEx(ctx, &models.MinAmountParams{
	CurrencyFrom:   "btc",
	CurrencyTo:     "usdttrc20",
	FiatEquivalent: "usd",
})
```

---

## Webhooks

NOWPayments posts an IPN to your `ipn_callback_url` each time a payment changes status. Check the signature against the raw body before you read it:

```go
import "github.com/muxover/nowpay-go/webhook"

func handleIPN(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !webhook.VerifySignature(body, r.Header.Get(webhook.SignatureHeader), ipnSecret) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	ev, err := webhook.ParseEvent(body)
	if err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}
	switch ev.Type {
	case webhook.EventPaymentFinished:
		// Fulfil ev.OrderID. The money has reached your wallet.
	case webhook.EventPaymentPartiallyPaid:
		// The customer sent less than the price.
	case webhook.EventPaymentFailed:
		// The payment failed or expired.
	}
	w.WriteHeader(http.StatusOK)
}
```

| Status | `ev.Type` |
|--------|-----------|
| `waiting`, `confirming` | `payment_created` |
| `confirmed`, `sending` | `payment_confirmed` |
| `partially_paid` | `payment_partially_paid` |
| `finished` | `payment_finished` |
| `failed`, `expired` | `payment_failed` |
| `refunded` | `payment_refunded` |
| any other status | empty |

`ev.AsPaymentEvent()` returns every field of the callback, including the fee breakdown, and `ev.AsInvoiceEvent()` its invoice fields. `webhook.Sign(body, secret)` produces a valid signature, which is handy for testing your handler locally.

---

## Error handling

Every 4xx and 5xx response comes back as a `*nowpay.APIError`, which matches the sentinel errors for its status:

```go
_, err := client.Payments.Get(ctx, id)
switch {
case errors.Is(err, nowpay.ErrPaymentNotFound):
	// 404 for this payment
case errors.Is(err, nowpay.ErrRateLimited):
	// 429
}

var apiErr *nowpay.APIError
if errors.As(err, &apiErr) {
	log.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
}
```

| Status | Sentinels |
|--------|-----------|
| 400 | `ErrBadRequest` |
| 401 | `ErrUnauthorized`, `ErrInvalidAPIKey` |
| 403 | `ErrForbidden` (also sent when the calling IP is not whitelisted; `Code` is `ENDPOINT_NOT_ALLOWED`) |
| 404 | `ErrNotFound`, plus `ErrPaymentNotFound`, `ErrInvoiceNotFound`, or `ErrSubscriptionNotFound` from the matching service |
| 422 | `ErrValidation` |
| 429 | `ErrRateLimited` |
| 5xx | `ErrServerError` |

---

## Examples

- [Create payment](examples/create_payment): `API_KEY=... go run ./examples/create_payment`
- [Webhook server](examples/webhook_server): `IPN_SECRET=... go run ./examples/webhook_server`

---

## Testing

```bash
go test -race ./...
```

The tests run every service against the example responses from the official NOWPayments API docs (`testdata/`) and against real responses recorded from the live API (`testdata/recorded/`). The webhook tests check signatures against ones made by Node.js with `JSON.stringify`, the same way NOWPayments signs them.

To run the same checks against your own account, see [live tests](CONTRIBUTING.md#live-tests). They are read-only on a production account unless you opt in to creating test records.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

Licensed under the [MIT](LICENSE) license.

---

## Links

- Repository: https://github.com/muxover/nowpay-go
- Issues: https://github.com/muxover/nowpay-go/issues
- Changelog: [CHANGELOG.md](CHANGELOG.md)
- Go Reference: https://pkg.go.dev/github.com/muxover/nowpay-go

---

<p align="center">Made with ❤️ by Jax (@muxover)</p>
