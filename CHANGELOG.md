# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/muxover/nowpay-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/muxover/nowpay-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/muxover/nowpay-go/releases/tag/v0.1.0
