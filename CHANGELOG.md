# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-06-20

Alignment pass against the official NOWPayments Postman collection. Contains breaking changes.

### Fixed

- **Base URL**: requests no longer hit a malformed `/v1/v1/...` path. `DefaultBaseURL` keeps the `/v1` segment and endpoint paths are now version-less (e.g. `/payment`).
- **Payouts — Create**: now posts the correct `{ withdrawals: [...] }` body to `POST /payout` (previously sent a single flat object). Removed the non-existent `POST /payout/batch` endpoint.
- **Payouts — Verify**: now `POST /payout/:batch-withdrawal-id/verify` with a `verification_code` body (was incorrectly a GET with no body).
- **Payouts — Get/List**: decode the real response envelopes (`GET /payout/:id` returns an array; `GET /payout` returns `{ "payouts": [...] }`).
- **Payments — List**: decode the `{ "data": [...] }` envelope (was decoding a bare array).
- **Payment model**: `expiration_estimate` → `expiration_estimate_date`; added `invoice_id`, `network`, `payin_hash`, `payout_hash`, `type`, `payment_extra_ids`.
- IDs and payout amounts now decode from both JSON numbers and quoted strings via `models.FlexInt` / `models.FlexFloat` (NOWPayments returns these inconsistently).

### Added

- **Conversions**: `Conversions.Create`, `Get`, `List` (`/conversion`).
- **Customers (Custody / sub-partner)**: `Create`, `Balance`, `List`, `Payments`, `Transfers`, `GetTransfer`, `Transfer`, `Deposit`, `WriteOff`, `DepositWithPayment`.
- **Fiat payouts**: `Providers`, `FiatCurrencies`, `CryptoCurrencies`, `PaymentMethods`, `CreateAccount`, `Accounts`, `Request`, `List` (`/fiat-payouts`).
- **Payouts**: `CancelBatch` (`POST /payout/:batch_id/cancel-batch`); scheduled-payout fields `ExecuteAt`/`Interval` and per-withdrawal `ipn_callback_url`, `fiat_amount`, `fiat_currency`, `unique_external_id`.
- **Payments**: `CreatePaymentRequest.IsFeePaidByUser`.
- **Webhooks**: `parent_payment_id` and `actually_paid` on parsed payment events.

### Changed (breaking)

- **Subscriptions** rewritten to match the real API: plural `/subscriptions` with a plan-based model. New methods: `CreatePlan`, `UpdatePlan`, `GetPlan`, `ListPlans`, `Create`, `Get`, `List`, `Delete`. The old price-based singular `/subscription` methods are removed.
- **Payouts.Create** now takes `*CreatePayoutRequest` (withdrawals) and returns `*PayoutBatch`; `Get` returns `[]Payout`; `Verify` takes a verification code; `BatchCreate` removed (use `Create`).
- Removed endpoints that do not exist in the API: `Payments.GetFlow`, `Payments.Refund`, `Currencies.Available`, `Invoices.Get` (and the `PaymentFlow`, `RefundRequest`, `RefundResponse` models).

## [0.2.0] - 2025-03-13

### Added

- **Auth**: `Client.Auth(ctx, email, password)` for JWT; optional `Config.Token` for List payments / Create payout
- **Balance**: `Client.Balance(ctx)` — balance per currency (GET /v1/balance)
- **Currencies**: `FullCurrencies` (GET /v1/full-currencies), `MerchantCoins` (GET /v1/merchant/coins), `SupportedWithFixedRate(ctx, bool)`
- **Payments**: `CreateFromInvoice` (POST /v1/invoice-payment), `UpdateMerchantEstimate` (POST /v1/payment/:id/update-merchant-estimate), `ListPaymentsParams.InvoiceID`
- **Estimates**: `MinAmountEx(ctx, params)` with fiat_equivalent, is_fixed_rate, is_fee_paid_by_user; extended `MinAmountResponse`
- **Invoices**: `CreateInvoiceRequest` extended with PayCurrency, PartiallyPaidURL, IsFixedRate, IsFeePaidByUser
- **Payouts**: `ValidateAddress`, `List`, `MinAmountForWithdrawal(coin)`, `Fee(currency, amount)`, `Cancel(withdrawalID)`, `Verify(batchWithdrawalID)`
- Models: `AuthRequest`/`AuthResponse`, `BalanceResponse`, `CreateInvoicePaymentRequest`, `UpdateMerchantEstimateResponse`, `MinAmountParams`, `ValidateAddressRequest`, `ListPayoutsParams`

### Changed

- Internal HTTP client supports optional JWT (Authorization header) when `Config.Token` is set
- **Retry**: optional `Config.Retry` (MaxRetries, InitialBackoff, MaxBackoff); retries on 429 and 5xx with exponential backoff for production use

## [0.1.0] - 2025-03-13

### Added

- Full NOWPayments API client: payments, invoices, currencies, estimates, payouts, subscriptions
- API status, payment flow, refunds, batch (mass) payouts, subscription cancel/update
- Webhook signature verification and typed event parsing
- Price conversion and fiat (USD, EUR) support in estimates and models
- Comprehensive error types and HTTP layer with context support
- Examples: create payment, webhook server

[Unreleased]: https://github.com/muxover/nowpay-go/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/muxover/nowpay-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/muxover/nowpay-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/muxover/nowpay-go/releases/tag/v0.1.0
