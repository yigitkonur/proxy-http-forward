# proxy-http-forward - Makefile
# High-performance HTTP/HTTPS proxy server in pure Go

BINARY_NAME := proxy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet
GOFMT := gofmt

# Build flags
LDFLAGS := -ldflags "-s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(BUILD_DATE)"

# Directories
CMD_DIR := ./cmd/proxy
BUILD_DIR := ./build
DIST_DIR := ./dist

.PHONY: all build clean test fmt vet lint run docker help

# Default target
all: clean fmt vet test build

## Build targets
build: ## Build the binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build for Linux (amd64)
	@echo "Building for Linux amd64..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

build-darwin: ## Build for macOS (arm64)
	@echo "Building for macOS arm64..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)

build-all: build-linux build-darwin ## Build for all platforms

## Development targets
run: build ## Build and run the proxy
	@echo "Starting proxy server..."
	$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode with hot reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

## Testing targets
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## Code quality targets
fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

lint: ## Run golangci-lint (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

## Dependencies targets
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify

## Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t proxy-http-forward:$(VERSION) .
	docker tag proxy-http-forward:$(VERSION) proxy-http-forward:latest

docker-run: ## Run Docker container
	docker run -p 8080:8080 -p 9090:9090 proxy-http-forward:latest

docker-push: ## Push Docker image (requires DOCKER_REGISTRY env var)
	@if [ -z "$(DOCKER_REGISTRY)" ]; then echo "DOCKER_REGISTRY not set"; exit 1; fi
	docker tag proxy-http-forward:$(VERSION) $(DOCKER_REGISTRY)/proxy-http-forward:$(VERSION)
	docker push $(DOCKER_REGISTRY)/proxy-http-forward:$(VERSION)

## Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out coverage.html

version: ## Show version
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(BUILD_DATE)"

## Help target
help: ## Show this help
	@echo "proxy-http-forward - Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Example usage:"
	@echo "  make build     # Build the binary"
	@echo "  make run       # Build and run"
	@echo "  make test      # Run tests"
	@echo "  make docker-build  # Build Docker image"
