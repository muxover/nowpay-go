# NowPay Go

<div align="center">

[![CI](https://github.com/muxover/nowpay-go/actions/workflows/ci.yml/badge.svg)](https://github.com/muxover/nowpay-go/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/muxover/nowpay-go.svg)](https://pkg.go.dev/github.com/muxover/nowpay-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/muxover/nowpay-go)](https://goreportcard.com/report/github.com/muxover/nowpay-go)

**Go SDK for NOWPayments — payments, fiat, webhooks, Telegram.**

</div>

---

NowPay Go is a production-ready Go client for the [NOWPayments](https://nowpayments.io) cryptocurrency payment gateway. Use it in backends, bots, and services to create payments, invoices, estimates (including USD/EUR and other fiat), payouts, and subscriptions, with webhook verification and Telegram bot helpers.

## Features

- Full NOWPayments REST API: payments, invoices, currencies, estimates, payouts, subscriptions
- Price conversion and fiat: USD, EUR and other fiat in estimates and models
- Webhook signature verification and strongly typed event parsing
- Telegram helpers: payment buttons, invoice messages
- Context support, configurable HTTP client, structured errors
- Minimal dependencies, idiomatic Go

## Installation

```bash
go get github.com/muxover/nowpay-go
```

## Quick Start

```go
package main

import (
	"context"
	"github.com/muxover/nowpay-go"
	"github.com/muxover/nowpay-go/models"
)

func main() {
	client := nowpay.NewClient(nowpay.Config{
		APIKey: "YOUR_API_KEY",
	})
	ctx := context.Background()

	payment, err := client.Payments.Create(ctx, &models.CreatePaymentRequest{
		PriceAmount:   10.00,
		PriceCurrency: "usd",
		PayCurrency:   "btc",
		OrderID:       "order-001",
	})
	if err != nil {
		panic(err)
	}
	// use payment.PaymentID, payment.PayAddress, payment.PayAmount, etc.
}
```

## API overview

| Module | Methods |
|--------|--------|
| **Payments** | `Create`, `Get`, `List` |
| **Invoices** | `Create`, `Get` |
| **Currencies** | `Supported`, `Available` |
| **Estimates** | `Estimate`, `EstimateByPrice`, `MinAmount` (crypto and fiat: USD, EUR, etc.) |
| **Payouts** | `Create`, `Get` |
| **Subscriptions** | `Create`, `Get` |

All methods take `context.Context` as the first argument.

## Price conversion and fiat

Use `client.Estimates` for amounts in USD, EUR, or other fiat:

```go
est, err := client.Estimates.Estimate(ctx, 25.0, "usd", "btc")
// est.EstimatedAmount is the crypto amount

min, err := client.Estimates.MinAmount(ctx, "usd", "btc")
```

Payment and estimate models include `PriceCurrency`, `PriceAmount`, and API-provided fiat fields where applicable.

## Webhook setup

Verify IPN callbacks and parse events:

```go
import "github.com/muxover/nowpay-go/webhook"

// In your HTTP handler:
body, _ := io.ReadAll(r.Body)
sig := r.Header.Get(webhook.SignatureHeader)
if !webhook.VerifySignature(body, sig, ipnSecret) {
	http.Error(w, "invalid signature", 401)
	return
}
ev, err := webhook.ParseEvent(body)
// ev.PaymentID, ev.PaymentStatus, ev.OrderID, etc.
// Use ev.AsPaymentEvent() or ev.AsInvoiceEvent() for typed payloads
```

Event types include payment_created, payment_confirmed, payment_finished, payment_failed, and invoice/payout/subscription events.

## Telegram bot example

Build payment messages and inline keyboards for Telegram:

```go
import (
	"github.com/muxover/nowpay-go"
	"github.com/muxover/nowpay-go/telegram"
)

payment, _ := client.Payments.Create(ctx, req)
msg := telegram.InvoiceMessage(payment)
kb := telegram.PaymentKeyboard("Pay with crypto", paymentURL)
// Send msg and kb to the user (e.g. with telegram-bot-api or similar)
```

See [examples/telegram_bot](examples/telegram_bot) for a minimal flow.

## Examples

- [Create payment](examples/create_payment) — create a payment and print result
- [Telegram bot](examples/telegram_bot) — payment + message and keyboard
- [Webhook server](examples/webhook_server) — verify IPN and parse events

Run with required env vars (e.g. `API_KEY`, `IPN_SECRET` for webhook).

## Configuration

| Field | Description |
|-------|-------------|
| `APIKey` | Required. Your NOWPayments API key. |
| `BaseURL` | Optional. Default `https://api.nowpayments.io/v1`. |
| `Timeout` | Optional. HTTP timeout (default 30s). |
| `HTTPClient` | Optional. Custom `*http.Client`. |

## Error handling

The SDK returns errors that wrap sentinel errors. Use `errors.Is` to check:

```go
if errors.Is(err, nowpay.ErrInvalidAPIKey) { ... }
if errors.Is(err, nowpay.ErrPaymentNotFound) { ... }
if errors.Is(err, nowpay.ErrRateLimited) { ... }
if errors.Is(err, nowpay.ErrInvalidSignature) { ... }
```

Available: `ErrInvalidAPIKey`, `ErrPaymentNotFound`, `ErrInvoiceNotFound`, `ErrPayoutNotFound`, `ErrSubscriptionNotFound`, `ErrRateLimited`, `ErrInvalidSignature`, `ErrBadRequest`, `ErrValidation`, `ErrServerError`, `ErrNotFound`. Use `nowpay.IsAPIError(err)` to detect any API error.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Licensed under [MIT](LICENSE).

## Links

- Repository: [github.com/muxover/nowpay-go](https://github.com/muxover/nowpay-go)
- Issues: [github.com/muxover/nowpay-go/issues](https://github.com/muxover/nowpay-go/issues)
- Changelog: [CHANGELOG.md](CHANGELOG.md)

---

<p align="center">Made with ❤️ by Jax (@muxover)</p>
