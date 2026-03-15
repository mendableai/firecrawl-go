# Contributing to firecrawl-go

Thank you for your interest in contributing!

## Quick Start

```bash
git clone git@github.com:firecrawl/firecrawl-go.git
cd firecrawl-go
go mod download
make check  # lint + vet + test
```

## Development Workflow

1. Fork the repository and create a feature branch from `main`.
2. Make your changes following the code style below.
3. Run `make check` before committing (lint + vet + unit tests).
4. Push and open a pull request with a clear description of what changed and why.

A pre-commit hook runs `make check` automatically on every commit.

## Code Style

- Format with `gofumpt`: `make fmt`
- Lint with `golangci-lint` v2: `make lint`
- Vet with `go vet`: `make vet`
- All public methods require `context.Context` as the first parameter.
- Optional request fields use pointer types with `json:",omitempty"`.
- Use typed request structs (internal, unexported) with `json.Marshal` for POST endpoints.
- Follow conventional commit format: `feat(scope): description`, `fix(scope): description`, `docs: description`.

## Testing

| Command | What It Runs | API Key? |
|---------|-------------|----------|
| `make test` | 160 unit tests (httptest mocks) | No |
| `make test-integration` | 32 E2E tests (live Firecrawl API) | Yes |
| `make coverage` | HTML coverage report | No |

Unit tests run against `httptest.NewServer` mock servers — no `.env` file or API key needed. If unit tests fail, the issue is in the code, not missing credentials.

For integration tests:

```bash
cp .env.example .env
# Edit .env:
#   API_URL=https://api.firecrawl.dev
#   TEST_API_KEY=fc-your-api-key
make test-integration
```

Integration tests consume API credits.

## Prerequisites

| Tool | Version | Installation |
|------|---------|-------------|
| Go | 1.23+ | [go.dev/dl](https://go.dev/dl/) |
| golangci-lint | v2.x | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest` |
| gofumpt | latest | `go install mvdan.cc/gofumpt@latest` |

## Adding a New Endpoint

1. Define request/response types in `types.go` with full godoc comments.
2. Create a new file `<endpoint>.go` with the public method(s).
3. Add a corresponding `<endpoint>_test.go` with unit tests using `httptest.NewServer`.
4. If the endpoint is async with polling, add E2E tests in `firecrawl_test.go` (build tag: `integration`).
5. Run `make check` to verify everything passes.

Every exported symbol must have a godoc comment. Public methods must document all parameters, return values, and any error conditions.

## Detailed Guide

For a comprehensive architecture overview, code patterns, request flow diagrams, and FAQ, see the [full contribution guide](../../specs/firecrawl-go-v2/contribution-guide.md) in the Agentic Layer specs.
