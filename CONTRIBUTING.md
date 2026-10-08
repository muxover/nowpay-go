# Contributing

NowPay Go is a small, dependency-free Go client for the NOWPayments API. Fixes for endpoints that behave differently from the docs are especially welcome. A new method needs a passing live test before it ships.

## Getting started

```bash
git clone https://github.com/muxover/nowpay-go.git
cd nowpay-go
go build ./...
```

Go 1.21 or later is required.

## Running tests

```bash
go vet ./...
go test -race ./...
```

The tests need no API key. `client_test.go` runs every service against an `httptest` server that returns the example responses stored in `testdata/`, which come from the official NOWPayments API collection. When you add or fix an endpoint, add its real response to `testdata/` and a case to `TestEndpoints`.

Webhook signature tests use `testdata/ipn_signatures.json`, signed with Node.js `crypto` and `JSON.stringify`, the way NOWPayments signs callbacks.

### Live tests

`live_test.go` runs every service against a real account and is left out of normal builds. It defaults to the sandbox; set `NOWPAYMENTS_BASE_URL=https://api.nowpayments.io/v1` for a production account.

```bash
NOWPAYMENTS_API_KEY=... NOWPAYMENTS_EMAIL=... NOWPAYMENTS_PASSWORD=... \
  go test -tags live -run Live -v -count=1 .
```

| Variable | Effect |
|----------|--------|
| `NOWPAYMENTS_EMAIL`, `NOWPAYMENTS_PASSWORD` | Unlock the endpoints that need a JWT. |
| `NOWPAYMENTS_LIVE_CREATE=1` | On production, also create a test payment, invoice, plan, and custody customer. The plan and customer are reused on later runs. |
| `NOWPAYMENTS_LIVE_STRICT=1` | Fail instead of skip when a call is refused or cannot run. A release needs a strict run to pass. |
| `NOWPAYMENTS_LIVE_DUMP` | Folder for the raw responses, `testdata/live/` by default (ignored by git, JWT redacted). |

Nothing in the live tests moves funds. Compare the raw responses with `testdata/recorded/` and add a sanitized copy when the API sends something new.

## Code style

- Format with `gofmt`. CI also runs `staticcheck` and markdownlint:

  ```bash
  go run honnef.co/go/tools/cmd/staticcheck@latest ./...
  npx markdownlint-cli2 "**/*.md"
  ```

- Keep the module free of third-party dependencies.
- Keep public types backward compatible. When the API returns a field as a number or a string, decode both (see `models.FlexInt`) rather than changing the field's type.

## Submitting changes

For anything bigger than a small fix, open an issue first. Branch from `main` (`fix/...` or `feat/...`), keep one change per PR, and add an entry under `[Unreleased]` in `CHANGELOG.md`.

## Reporting bugs

Use the issue templates. Include the Go version, the method you called, the error, and the API response body if you have it (remove any keys or addresses first).
