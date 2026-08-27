.PHONY: all fmt lint test test-integration coverage check build install-hooks clean

BIN_DIR := bin
BINARY_NAME := zentao

all: check build

fmt:
	@echo "==> Formatting Go files with gofmt..."
	gofmt -s -w .

lint:
	@echo "==> Running golangci-lint (format, security, complexity & static checks)..."
	golangci-lint run ./...

test:
	@echo "==> Running unit tests with race detector..."
	go test -race -v ./...

test-integration:
	@echo "==> Running full end-to-end integration tests (Zero-Pollution)..."
	go test -race -v -tags=integration -run=TestIntegration_FullLifecycle ./cmd/...

coverage:
	@echo "==> Running full test suite and generating code coverage..."
	go test -race -coverpkg=./pkg/zentao,./internal/...,./cmd -tags=integration -coverprofile=coverage.out ./cmd/...
	@echo ""
	@echo "==================== COVERAGE SUMMARY ===================="
	go tool cover -func=coverage.out
	@echo "=========================================================="
	@go tool cover -html=coverage.out -o coverage.html
	@echo "==> HTML coverage report generated at coverage.html"

check: lint test
	@echo "==> All quality and race checks passed!"

build:
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) main.go

install-hooks:
	@echo "==> Installing Git hooks (pre-commit, reference-transaction, pre-push)..."
	chmod +x .githooks/*
	git config core.hooksPath .githooks
	@echo "==> Git hooks configured successfully!"

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(BINARY_NAME) dist coverage.out coverage.html
