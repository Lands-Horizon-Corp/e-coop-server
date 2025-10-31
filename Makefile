# E-Coop Server Makefile
# A comprehensive build system for the financial cooperative management system

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run
BINARY_NAME=e-coop-server
MAIN_PATH=main.go

# Docker parameters
DOCKER_COMPOSE=docker compose

# Default target
.PHONY: help
help: ## Show this help message
	@echo "E-Coop Server - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development Commands
.PHONY: dev
dev: ## Start development server
	$(GORUN) $(MAIN_PATH) server

.PHONY: build
build: ## Build the application binary
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

.PHONY: clean
clean: ## Clean build artifacts and caches
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	$(GORUN) $(MAIN_PATH) cache clean

# Database Commands
.PHONY: db-migrate
db-migrate: ## Migrate database schema
	$(GORUN) $(MAIN_PATH) db migrate

.PHONY: db-seed
db-seed: ## Seed database with initial data
	$(GORUN) $(MAIN_PATH) db seed

.PHONY: db-reset
db-reset: ## Reset database (drops all tables)
	$(GORUN) $(MAIN_PATH) db reset

.PHONY: db-refresh
db-refresh: ## Reset database and seed with fresh data
	$(GORUN) $(MAIN_PATH) db refresh

.PHONY: db-setup
db-setup: db-migrate db-seed ## Complete database setup (migrate + seed)

# Cache Commands
.PHONY: cache-clean
cache-clean: ## Clean application cache
	$(GORUN) $(MAIN_PATH) cache clean

# Docker Commands
.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	$(DOCKER_COMPOSE) up --build -d

.PHONY: docker-down
docker-down: ## Stop all Docker services
	$(DOCKER_COMPOSE) down

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart all Docker services

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	$(DOCKER_COMPOSE) logs -f

# Testing Commands
.PHONY: test
test: ## Run all tests
	$(GOTEST) -v ./services/horizon_test

.PHONY: test-clean
test-clean: ## Run tests with clean cache
	$(GOCLEAN) -cache
	$(GOTEST) -v ./services/horizon_test

# Code Quality Commands
.PHONY: format
format: ## Format code with goimports and gofmt
	goimports -w .
	gofmt -w .

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: quality
quality: format lint ## Run all code quality checks

# Environment Commands
.PHONY: env-setup
env-setup: ## Setup environment file
	cp .env.example .env
	@echo "Please edit .env file with your configuration"

# Port Management
.PHONY: kill-ports
kill-ports: ## Kill processes using conflicting ports
	chmod +x kill_ports.sh
	./kill_ports.sh

# Dependencies
.PHONY: deps
deps: ## Download and tidy dependencies
	$(GOGET) -d ./...
	$(GOMOD) tidy

.PHONY: deps-update
deps-update: ## Update all dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Installation Commands
.PHONY: install
install: build ## Install binary to system
	sudo cp $(BINARY_NAME) /usr/local/bin/

.PHONY: uninstall
uninstall: ## Remove binary from system
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# Development Setup
.PHONY: setup
setup: env-setup deps docker-up db-setup ## Complete development environment setup

# Production Build
.PHONY: build-prod
build-prod: ## Build production binary
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BINARY_NAME) $(MAIN_PATH)

# Deployment Commands
.PHONY: deploy-check
deploy-check: quality test ## Pre-deployment checks
	@echo "All checks passed! Ready for deployment."

.PHONY: deploy-fly
deploy-fly: deploy-check ## Deploy to Fly.io
	fly deploy
	fly machine restart 148e4d55f36278
	fly machine restart 90802d3ea0ed38

.PHONY: deploy-logs
deploy-logs: ## Show deployment logs
	fly logs

# Quick Commands
.PHONY: start
start: docker-up db-setup dev ## Quick start (setup + run)

.PHONY: reset
reset: docker-down clean docker-up db-refresh dev ## Complete reset and restart

# Utility Commands
.PHONY: version
version: ## Show version information
	$(GORUN) $(MAIN_PATH) version

.PHONY: routes
routes: ## Show available API routes (requires server to be running)
	@echo "API routes available at: http://localhost:8000/routes"
	@echo "Make sure the server is running first!"

# Advanced Commands
.PHONY: benchmark
benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: coverage
coverage: ## Generate test coverage report
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: mod-graph
mod-graph: ## Show module dependency graph
	$(GOMOD) graph

# Clean everything
.PHONY: clean-all
clean-all: clean docker-down ## Clean everything (build artifacts, Docker, cache)
	docker system prune -f
	$(GOCLEAN) -modcache
