.PHONY: build test clean run serve help install uninstall build-frontend dev-frontend

# Binary names
CLI_BINARY=conclave
SERVER_BINARY=conclave-server

# Build directory
BUILD_DIR=bin
WEB_DIR=web/app

# Go parameters
# Use mise-managed Go if available (fixes GOROOT mismatch)
MISE_GO_ROOT := $(shell mise where go 2>/dev/null)
ifdef MISE_GO_ROOT
  export GOROOT := $(MISE_GO_ROOT)
endif
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get

# Build flags
LDFLAGS=-ldflags "-s -w"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: build-frontend ## Build CLI and server binaries with frontend
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/conclave/
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BINARY) ./cmd/server/

build-cli: ## Build only CLI binary
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/conclave/

build-server: build-frontend ## Build only server binary
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
	rm -rf $(WEB_DIR)/dist

build-frontend: ## Build React frontend
	@echo "Building React frontend..."
	@cd $(WEB_DIR) && npm run build
	@echo "Frontend built successfully"

dev-frontend: ## Run React frontend in development mode
	@cd $(WEB_DIR) && npm run dev

run: build-cli ## Build and show CLI help
	./$(BUILD_DIR)/$(CLI_BINARY) --help

serve: build-server ## Build and start web server
	./$(BUILD_DIR)/$(SERVER_BINARY)

install: build ## Install conclave to ~/.local/bin (no sudo)
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(CLI_BINARY) ~/.local/bin/
	@echo "Installed $(CLI_BINARY) to ~/.local/bin/"
	@echo "Ensure ~/.local/bin is in your PATH"

uninstall: ## Remove conclave from common locations
	@rm -f /usr/local/bin/$(CLI_BINARY) 2>/dev/null || true
	@rm -f ~/.local/bin/$(CLI_BINARY) 2>/dev/null || true
	@rm -f $(GOPATH)/bin/$(CLI_BINARY) 2>/dev/null || true
	@rm -f ~/go/bin/$(CLI_BINARY) 2>/dev/null || true
	@echo "Uninstalled $(CLI_BINARY)"

deps: ## Download dependencies
	$(GOGET) -v ./...

lint: ## Run linter (requires golangci-lint)
	golangci-lint run ./...

# Development helpers
dev-serve: ## Run server with auto-reload (requires air)
	air -c .air.toml

.DEFAULT_GOAL := help
