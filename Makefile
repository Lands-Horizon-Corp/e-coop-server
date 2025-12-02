# E-Coop Server Makefile
# Based on cmd/actions.go functionality

.PHONY: help build clean run server test docker compose

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := e-coop-server
BUILD_DIR := ./bin
GO_FILES := $(shell find . -type f -name '*.go')

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## Help
help: ## Show this help message
	@echo "$(BLUE)E-Coop Server Makefile$(NC)"
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Build
build: ## Build the application binary
	@echo "$(BLUE)Building application...$(NC)"
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-linux: ## Build for Linux
	@echo "$(BLUE)Building for Linux...$(NC)"
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux .
	@echo "$(GREEN)Linux build completed: $(BUILD_DIR)/$(BINARY_NAME)-linux$(NC)"

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean
	@echo "$(GREEN)Clean completed$(NC)"

go-clean-caches: ## Clean all Go caches (build, mod, test, fuzz)
	@echo "$(YELLOW)Cleaning Go caches...$(NC)"
	go clean -cache -modcache -testcache -fuzzcache
	@echo "$(GREEN)Go caches cleaned$(NC)"

## Development
dev: ## Run in development mode (build and run server)
	@echo "$(BLUE)Starting development server...$(NC)"
	go run . server

run: dev ## Alias for dev

## Server Management
server: build ## Build and start the server
	@echo "$(BLUE)Starting server...$(NC)"
	$(BUILD_DIR)/$(BINARY_NAME) server

## Database Commands
db-migrate: ## Migrate database schema
	@echo "$(BLUE)Migrating database...$(NC)"
	go run . db migrate

db-seed: ## Seed database with initial data
	@echo "$(BLUE)Seeding database...$(NC)"
	go run . db seed

db-reset: ## Reset database (drops and recreates)
	@echo "$(YELLOW)Resetting database...$(NC)"
	go run . db reset

db-refresh: ## Reset and seed database
	@echo "$(BLUE)Refreshing database...$(NC)"
	go run . db refresh

db-seed: ## Run performance seed (default multiplier: 1)
	@echo "$(BLUE)Running performance seed...$(NC)"
	go run . db performance-seed


cache-clean:
	@echo "$(YELLOW)Cleaning cache...$(NC)"
	go run . cache clean

## Security Commands
security-enforce: ## Update HaGeZi blocklist
	@echo "$(BLUE)Enforcing HaGeZi blocklist...$(NC)"
	go run . security enforce

security-clear: ## Clear all blocked IPs from cache
	@echo "$(YELLOW)Clearing blocked IPs...$(NC)"
	go run . security clear

## Setup Commands
setup: db-migrate db-seed
	@echo "$(GREEN)Setup completed$(NC)"

fresh-start: db-refresh cache-clean
	@echo "$(GREEN)Fresh start completed$(NC)"

full-reset-and-run: go-clean-caches tidy cache-clean security-enforce db-reset db-migrate db-seed server
	@echo "$(GREEN)Full reset and server start completed$(NC)"

hesoyam:
	@echo "$(BLUE)Pulling latest changes from git...$(NC)"
	git pull
	@echo "$(GREEN)Git pull completed$(NC)"
	@$(MAKE) full-reset-and-run
