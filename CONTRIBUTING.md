# Contributing

## Getting started

Clone the repository and build:

```bash
git clone https://github.com/muxover/nowpay-go.git
cd nowpay-go
go build ./...
```

## Running tests

```bash
go test ./...
go vet ./...
```

## Code style

- Use `gofmt` for formatting.
- Follow standard Go conventions and idiomatic style.

## Submitting changes

1. Open an issue to discuss larger changes.
2. Branch from `main` (e.g. `feat/your-feature` or `fix/your-fix`).
3. Make your changes; keep each PR focused on one change.
4. Ensure `go build ./...` and `go test ./...` pass.
5. Open a pull request with a clear description and reference any related issue.

## Reporting issues

Include: Go version, OS, steps to reproduce, and the full error message or behavior.
