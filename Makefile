.PHONY: build test clean run serve help

# Binary names
CLI_BINARY=dbate
SERVER_BINARY=dbate-server

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

# Build flags
LDFLAGS=-ldflags "-s -w"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build CLI and server binaries
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/dbate/
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BINARY) ./cmd/server/

build-cli: ## Build only CLI binary
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/dbate/

build-server: ## Build only server binary
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BINARY) ./cmd/server/

test: ## Run all tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

run: build-cli ## Build and show CLI help
	./$(BUILD_DIR)/$(CLI_BINARY) --help

serve: build-cli ## Build and start web server
	./$(BUILD_DIR)/$(CLI_BINARY) serve

install: build ## Install binaries to $GOPATH/bin
	cp $(BUILD_DIR)/$(CLI_BINARY) $(GOPATH)/bin/
	@echo "Installed $(CLI_BINARY) to $(GOPATH)/bin/"

deps: ## Download dependencies
	$(GOGET) -v ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

# Development helpers
dev-serve: ## Run server with auto-reload (requires air)
	air -c .air.toml

.DEFAULT_GOAL := help
