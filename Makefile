.PHONY: build test clean run serve help install install-user install-gopath uninstall

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

install: build ## Install dbate to /usr/local/bin (may require sudo)
	@if [ -w /usr/local/bin ]; then \
		cp $(BUILD_DIR)/$(CLI_BINARY) /usr/local/bin/; \
		echo "Installed $(CLI_BINARY) to /usr/local/bin/"; \
	else \
		echo "Installing to /usr/local/bin requires sudo"; \
		sudo cp $(BUILD_DIR)/$(CLI_BINARY) /usr/local/bin/; \
		echo "Installed $(CLI_BINARY) to /usr/local/bin/"; \
	fi

install-user: build ## Install dbate to ~/.local/bin (no sudo)
	@mkdir -p ~/.local/bin
	cp $(BUILD_DIR)/$(CLI_BINARY) ~/.local/bin/
	@echo "Installed $(CLI_BINARY) to ~/.local/bin/"
	@echo "Ensure ~/.local/bin is in your PATH"

install-gopath: build ## Install dbate to GOPATH/bin
	@if [ -z "$(GOPATH)" ]; then \
		echo "GOPATH not set, using ~/go/bin"; \
		mkdir -p ~/go/bin; \
		cp $(BUILD_DIR)/$(CLI_BINARY) ~/go/bin/; \
		echo "Installed $(CLI_BINARY) to ~/go/bin/"; \
	else \
		cp $(BUILD_DIR)/$(CLI_BINARY) $(GOPATH)/bin/; \
		echo "Installed $(CLI_BINARY) to $(GOPATH)/bin/"; \
	fi

uninstall: ## Remove dbate from common locations
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
