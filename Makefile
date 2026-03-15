.DEFAULT_GOAL := help
.PHONY: help build test test-integration lint fmt vet coverage clean check

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compile the library
	go build ./...

test: ## Run unit tests (no API key needed)
	go test -race -v -count=1 ./...

test-integration: ## Run integration tests (requires .env with API key)
	go test -race -v -count=1 -tags=integration ./...

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format code with gofumpt
	gofumpt -w .

vet: ## Run go vet
	go vet ./...

coverage: ## Generate HTML coverage report
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Remove generated files
	rm -f coverage.out coverage.html

check: lint vet test ## Run all checks (lint + vet + test)
