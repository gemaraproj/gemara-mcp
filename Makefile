# SPDX-License-Identifier: Apache-2.0

.PHONY: build test test-integration vet fmt lint golangci-lint clean help

# Binary name
BINARY_NAME := gemara-mcp

# Build directory
BUILD_DIR := bin

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOFMT := $(GOCMD) fmt

# Version information
# Try to get version from git tag, fallback to commit hash or default
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null)
GIT_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)

VERSION ?= $(if $(GIT_TAG),$(GIT_TAG),$(if $(GIT_VERSION),$(GIT_VERSION),0.1.0))
BUILD ?= $(if $(GIT_COMMIT),$(GIT_COMMIT),dev)
VERSION_PKG := github.com/gemaraproj/gemara-mcp/internal/cli

# Build flags
LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Build=$(BUILD)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@echo "Version: $(VERSION)"
	@echo "Build: $(BUILD)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-integration: ## Run tests including integration (requires network)
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

fmt: ## Run go fmt
	@echo "Running go fmt..."
	$(GOFMT) ./...

fmt-check: ## Check if code is formatted (for CI)
	@echo "Checking code formatting..."
	@if command -v gofmt >/dev/null 2>&1; then \
		if [ $$(gofmt -l . | grep -v "^$$" | wc -l) -ne 0 ]; then \
			echo "Code is not formatted. Run 'make fmt' to fix."; \
			gofmt -d .; \
			exit 1; \
		fi; \
	else \
		echo "gofmt not found, skipping format check"; \
	fi
	@echo "Code is properly formatted."

golangci-lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
		exit 1; \
	fi

lint: vet fmt-check golangci-lint ## Run all linting checks

ci: fmt-check vet golangci-lint test ## Run all CI checks

compose-gateway: ## Start gateway auth stack (Envoy validates JWTs)
	docker compose -f hack/docker-compose.yml --profile gateway up --build

compose-hydra: ## Start Hydra DCR stack (MCP client OAuth testing)
	docker compose -f hack/docker-compose.yml --profile hydra up --build

compose-down: ## Stop all compose services
	docker compose -f hack/docker-compose.yml --profile gateway --profile hydra down -v

compose-token: ## Fetch a test token from the mock IdP
	@curl -sf -X POST http://localhost:9090/default/token \
		-H "Content-Type: application/x-www-form-urlencoded" \
		-d "grant_type=client_credentials&client_id=test&client_secret=test&scope=openid" \
		| jq -r .access_token

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete."