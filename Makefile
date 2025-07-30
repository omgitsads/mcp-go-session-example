# MCP Go Session Example - Makefile
# Provides common development and deployment commands

# Variables
BINARY_NAME=mcp-server
BUILD_DIR=./bin
CMD_DIR=./cmd
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

# Help target - shows available commands
.PHONY: help
help: ## Show this help message
	@echo "MCP Go Session Example - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
.PHONY: dev
dev: ## Start the server in development mode (with hot reload using air if available)
	@if command -v air >/dev/null 2>&1; then \
		echo "Starting server with hot reload..."; \
		air; \
	else \
		echo "Hot reload not available. Install 'air' with: go install github.com/air-verse/air@latest"; \
		echo "Starting server normally..."; \
		go run $(CMD_DIR) server; \
	fi

.PHONY: run
run: ## Run the server (requires Redis at localhost:6379)
	REDIS_ADDR=localhost:6379 go run $(CMD_DIR) server

# Build commands
.PHONY: build
build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

.PHONY: build-all
build-all: ## Build binaries for multiple platforms
	@mkdir -p $(BUILD_DIR)
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Built binaries:"
	@ls -la $(BUILD_DIR)/

.PHONY: install
install: ## Install the binary to $GOPATH/bin
	go install $(LDFLAGS) $(CMD_DIR)

# Testing commands
.PHONY: test
test: ## Run all tests
	go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	go test -race -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bench
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Code quality commands
.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (install first: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

.PHONY: check
check: fmt vet test ## Run all code quality checks

# Dependency management
.PHONY: deps
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update: ## Update all dependencies
	go get -u ./...
	go mod tidy

# Docker commands (if using Docker)
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 8080:8080 --rm $(BINARY_NAME):$(VERSION)

.PHONY: docker-run-redis
docker-run-redis: ## Run Docker container with Redis
	docker run -p 8080:8080 -e REDIS_ADDR=host.docker.internal:6379 --rm $(BINARY_NAME):$(VERSION)

# Utility commands
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")"

# Redis development helpers
.PHONY: redis-start
redis-start: ## Start Redis in Docker for development
	@echo "Starting Redis container..."
	docker run --name mcp-redis -p 6379:6379 -d redis:7-alpine
	@echo "Redis started on localhost:6379"

.PHONY: redis-stop
redis-stop: ## Stop Redis Docker container
	@echo "Stopping Redis container..."
	docker stop mcp-redis && docker rm mcp-redis

.PHONY: redis-logs
redis-logs: ## Show Redis container logs
	docker logs -f mcp-redis

# Development workflow targets
.PHONY: setup
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reload..."; \
		go install github.com/cosmtrek/air@latest; \
	fi
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Development environment ready!"

.PHONY: ci
ci: check test-race test-cover ## Run all CI checks (format, vet, test with race detection and coverage)

# Server management
.PHONY: start
start: build ## Build and start the server (requires Redis)
	REDIS_ADDR=localhost:6379 $(BUILD_DIR)/$(BINARY_NAME) server

.PHONY: start-with-redis
start-with-redis: build ## Build and start the server with Redis
	REDIS_ADDR=localhost:6379 $(BUILD_DIR)/$(BINARY_NAME) server

# Configuration examples
.PHONY: config-example
config-example: ## Show configuration examples
	@echo "Environment Variable Examples:"
	@echo ""
	@echo "# Basic configuration"
	@echo "export MCP_HOST=0.0.0.0"
	@echo "export MCP_PORT=8080"
	@echo ""
	@echo "# Redis configuration"
	@echo "export REDIS_ADDR=localhost:6379"
	@echo "export REDIS_PASSWORD=mypassword"
	@echo "export REDIS_DB=1"
	@echo "export REDIS_PREFIX=myapp:mcp:"
	@echo "export REDIS_TTL=2h"
	@echo ""
	@echo "# Start server"
	@echo "make run"

# Force rebuild of targets that don't correspond to files
.PHONY: all
all: clean deps check build test ## Run a complete build pipeline