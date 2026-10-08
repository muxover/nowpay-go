# NowPay Go — maintainer notes

## What this is

NowPay Go is a Go client for the NOWPayments crypto payment API, for Go backends that take crypto payments. It covers payments, invoices, currencies, estimates, recurring payments, custody customers, and the payout, conversion, and fiat-payout lookups, and verifies and parses IPN callbacks. It uses only the standard library.

## Stack & layout

- Go 1.21+, no dependencies. Module `github.com/muxover/nowpay-go`.
- `client.go`: `Client`, `Config`, and every service. `internal/`: HTTP transport, retries, `APIError`. `models/`: request and response types, with `FlexInt` / `FlexFloat` for values the API sends as numbers or strings. `webhook/`: signature checks and callback parsing. `doc.go` files hold the package docs.
- `examples/`: `create_payment` and `webhook_server`.
- Tests: `client_test.go` runs every service against the example responses from the [NOWPayments API docs](https://documenter.getpostman.com/view/7907941/2s93JusNJt) in `testdata/`; `recorded_test.go` runs them against real responses in `testdata/recorded/`; `webhook/webhook_test.go` checks signatures against ones made by Node.js. `live_test.go` (`go test -tags live`) calls a real account from environment variables; it is read-only on production unless `NOWPAYMENTS_LIVE_CREATE=1`, and `NOWPAYMENTS_LIVE_STRICT=1` turns every skip into a failure.
- CI: `.github/workflows/ci.yml` — Go 1.21 and 1.27 with `-race`, gofmt, staticcheck, markdownlint. `.github/workflows/release.yml` runs CI on a `v*` tag, creates the GitHub release named from `release-notes/vX.Y.Z.md`, and asks proxy.golang.org to fetch the version.

## Decisions

- 2026-10-08: Balances decode an empty `[]` as an empty map, since the API sends that for a customer with no funds; a non-empty array is an error rather than being dropped.
- 2026-09-30: v1.0.0 is the first complete version and the only release; v0.1.0 through v0.5.0 are retracted in `go.mod`.
- 2026-09-30: A method ships only after a strict live run passes. Endpoints that move funds (payout create, verify and cancel; custody transfer, deposit and write-off; conversion create; fiat-payout account and request) are left out until they can be verified.
- 2026-09-30: The webhook package parses payment callbacks only, since the SDK does not create payouts.
- 2026-09-28: IPN signatures are checked by re-serializing the body like `JSON.stringify` (no HTML escaping). Arrays are also tried in the index-keyed form the docs' Node sample produces.
- 2026-09-28: Retries never repeat a `POST` after a 5xx, to avoid double payments. 429 is retried for every method and honors `Retry-After`.
- 2026-09-28: Fields the API sends as either numbers or strings keep their Go types and decode through custom `UnmarshalJSON`.
- 2026-09-28: `Payments.List` defaults to `created_at`, newest first; the API answers 500 when the sort fields are missing.
- 2026-09-28: A published version's files never change once proxy.golang.org has cached them. Code fixes go in a new version; docs-only changes stay on `main`.
- 2026-09-28: Every version gets release notes and a GitHub release. Commits and tags are signed with the maintainer's SSH signing key.

## Status

- v1.0.0 is the only version.
- Every method passes a strict live run against a production account, and is tested against the documented and the recorded responses.
- Not confirmed live: a real IPN callback (signing is checked against a Node.js reproduction).

## Next

- To release: add a CHANGELOG entry and `release-notes/vX.Y.Z.md`, sign the commit, push `main`, then push the tag.
- The fund-moving endpoints can be added in a minor version once each is verified on a sandbox account.

## Open questions

- None.
