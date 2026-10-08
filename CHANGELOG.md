# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-10-08

The first complete release. Every method was run against the production NOWPayments API and is covered by tests against both the documented and the recorded responses. Earlier v0.x versions are retracted.

### Added

- `Client` with `Status`, `Auth` (JWT from email and password), and `Balance`.
- `Payments`: `Create`, `CreateFromInvoice`, `Get`, `List` (newest first by default), and `UpdateMerchantEstimate`.
- `Invoices.Create`.
- `Currencies`: `Supported`, `SupportedWithFixedRate`, `FullCurrencies`, and `MerchantCoins`.
- `Estimates`: `Estimate`, `EstimateByPrice`, `MinAmount`, and `MinAmountEx`.
- `Payouts`: `List`, `ValidateAddress`, `MinAmountForWithdrawal`, and `Fee`.
- `Subscriptions`: plans (`CreatePlan`, `UpdatePlan`, `GetPlan`, `ListPlans`) and email subscriptions (`Create`, `Get`, `List`, `Delete`).
- `Conversions.List`.
- `Customers` (custody): `Create`, `Balance`, `List`, `Payments`, `Transfers`, and `DepositWithPayment`.
- `FiatPayouts`: `Providers`, `FiatCurrencies`, `CryptoCurrencies`, `PaymentMethods`, `Accounts`, and `List`.
- `webhook` package: `VerifySignature` and `Sign` match how NOWPayments signs IPN callbacks, and `ParseEvent` returns a typed payment event with the full body through `AsPaymentEvent` and `AsInvoiceEvent`.
- `APIError` with the status, NOWPayments error code, and message, matched by sentinel errors through `errors.Is`.
- Optional retries: 429 for every request, honoring `Retry-After`, and 5xx only for requests that are safe to repeat.
- IDs and amounts decode whether the API sends them as numbers or strings, and an empty balance decodes whether it arrives as `{}` or `[]`.
- `SandboxBaseURL`, a configurable `*http.Client`, and `context.Context` on every call.
- Opt-in live tests (`go test -tags live`) with a strict mode that fails on any skipped call.

[Unreleased]: https://github.com/muxover/nowpay-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/muxover/nowpay-go/releases/tag/v1.0.0
