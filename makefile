# ====================================================================================
# ENVIRONMENT CONFIG
# ====================================================================================

SHELL := /bin/bash
ENV_FILE := .env

ifneq (,$(wildcard $(ENV_FILE)))
	include $(ENV_FILE)
	export
endif

define load_env
	@bash -c '\
		if [ -f "$(ENV_FILE)" ]; then \
			echo "Loading env from $(ENV_FILE)..."; \
			set -o allexport; source "$(ENV_FILE)"; set +o allexport; \
			$(1); \
		else \
			echo "$(ENV_FILE) not found."; \
		fi'
endef

print-env:
	@$(call load_env, echo "APP_NAME=$$APP_NAME" && echo "APP_VERSION=$$APP_VERSION")

print-apmenv:
	@echo "ELASTIC_APM_SERVER_URL=$(ELASTIC_APM_SERVER_URL)"
	@echo "ELASTIC_APM_SECRET_TOKEN=$(ELASTIC_APM_SECRET_TOKEN)"
	@echo "ELASTIC_APM_SERVICE_NAME=$(ELASTIC_APM_SERVICE_NAME)"
	@echo "ELASTIC_APM_SERVICE_VERSION=$(ELASTIC_APM_SERVICE_VERSION)"
	@echo "ELASTIC_APM_SERVICE_NODE_NAME=$(ELASTIC_APM_SERVICE_NODE_NAME)"
	@echo "ELASTIC_APM_ENVIRONMENT=$(ELASTIC_APM_ENVIRONMENT)"

# ====================================================================================
# VARIABLES
# ====================================================================================

# Go variables
BINARY_NAME=go-app-temp
GO_VERSION ?= $(shell go version)

# Docker variables
DOCKER_COMPOSE=docker-compose


# ====================================================================================
# SETUP
# ====================================================================================

# .PHONY ensures that these targets are always run, even if a file with the same name exists.
.PHONY: all help docker-build up down logs ps docker-prune local-run local-build test mod-tidy mod-download

# Set the default command to run when `make` is called without arguments.
DEFAULT_GOAL := help


# ====================================================================================
# HELPERS
# ====================================================================================

help: ## ✨ Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)


# ====================================================================================
# CODE QUALITY
# ====================================================================================

lint: ## 🧐 Run golangci-lint to analyze source code
	@rm -rf ./reports/* 2>/dev/null || true
	@command -v golangci-lint >/dev/null 2>&1 || \
		(echo "--> golangci-lint not found. Please run 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest' to install." && exit 1)
	@echo "Running golangci-lint..."
	@golangci-lint run -v --fix --timeout=5m ./...


# ====================================================================================
# DOCKER WORKFLOW (for running the full stack)
# ====================================================================================

docker-build: ## 🐳 Build all Docker images for the project
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build

up: ## 🚀 Start all services in the background (Elastic, Kibana, App, etc.)
	@echo "Starting Docker containers in detached mode..."
	$(DOCKER_COMPOSE) up -d

down: ## 🛑 Stop and remove all running containers
	@echo "Stopping and removing Docker containers..."
	$(DOCKER_COMPOSE) down

full-down: ## 🗑️ Stop containers and remove volumes (deletes all Elasticsearch data)
	@echo "WARNING: This will delete all container data (e.g., Elasticsearch, Kibana)."
	$(DOCKER_COMPOSE) down -v

logs: ## 📜 Stream logs from all running containers
	@echo "Streaming logs from all services... (Press Ctrl+C to stop)"
	$(DOCKER_COMPOSE) logs -f

ps: ## 📊 Show the status of all running containers
	@echo "Current container status:"
	$(DOCKER_COMPOSE) ps

docker-prune: ## 🧹 Clean up unused Docker images, networks, and volumes
	@echo "Cleaning up dangling Docker resources..."
	docker system prune -a -f

run-standalone: ## 🧪 Build and run the app in a standalone Docker container
	@echo "Stopping and removing existing container (if any)..."
	docker stop go-app-temp || true
	docker rm go-app-temp || true
	@echo "Building standalone Docker image..."
	docker build -t go-app-temp .
	@echo "Running container on port 8090..."
	docker run -d --name go-app-temp --env-file ./.env -p 8090:8080 go-app-temp


# ====================================================================================
# LOCAL DEVELOPMENT WORKFLOW (for working on the Go app)
# ====================================================================================

local-run: ## 🏃 Run the Go application locally
	@echo "Starting the application locally..."
	go run main.go run -e local -d

local-build: ## 🛠️ Build the Go binary for your local machine
	@echo "Building Go binary for local environment..."
	go build -o $(BINARY_NAME) .
	@echo "Binary '$(BINARY_NAME)' created."

test: ## 🧪 Run all Go tests in the project
	@echo "Running Go tests..."
	go test -v ./...

migrate: ## 🛠️ Run database migrations using the migrate command
	@echo "Running database migrations..."
	go run main.go migrate -c .env
	@echo "Migration completed."

migrate-reset: ## 🛠️ Reset database migrations
	@echo "Resetting database migrations..."
	go run main.go migrate -c .env -r
	@echo "Migration reset completed."


# ====================================================================================
# GO MODULES MANAGEMENT
# ====================================================================================

mod-tidy: ## 🧹 Tidy up the go.mod and go.sum files
	@echo "Running go mod tidy..."
	go mod tidy

mod-download: ## 📥 Download Go module dependencies
	@echo "Downloading Go modules..."
	go mod download


# ====================================================================================
# DATABASE MANAGEMENT (MySQL)
# ====================================================================================

mysql-up: ## 🐘 Start the MySQL database container
	@echo "Starting MySQL container using .env configuration..."
	docker run --name my-mysql \
	  -e MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) \
	  -e MYSQL_DATABASE=$(MYSQL_DATABASE) \
	  -e MYSQL_USER=$(MYSQL_USER) \
	  -e MYSQL_PASSWORD=$(MYSQL_PASSWORD) \
	  -p 3306:3306 \
	  -d mysql:8

mysql-down: ## 🐘 Stop and remove the MySQL database container
	@echo "Stopping and removing MySQL container..."
	docker stop my-mysql || true
	docker rm my-mysql || true

mysql-logs: ## 🐘 View the logs of the MySQL container
	@echo "Following MySQL logs..."
	docker logs -f my-mysql

mysql-connect: ## 🐘 Connect to the MySQL database using the mysql client
	@echo "Connecting to MySQL database..."
	docker exec -it my-mysql mysql -u $(MYSQL_USER) -p$(MYSQL_PASSWORD) $(MYSQL_DATABASE)


# ====================================================================================
# DATABASE MANAGEMENT (TiDB)
# ====================================================================================

tidb-up: ## 🐯 Start the TiDB database container
	@echo "Starting TiDB container using default configuration..."
	docker run --name tidb-server \
	  -p 4000:4000 \
	  -p 10080:10080 \
	  -d pingcap/tidb:latest

	@echo "Waiting for TiDB to be ready..."
	@sleep 5

	@echo "Creating database '$(MYSQL_DB_NAME)' if not exists..."
	mysqlsh --sql -u root -h 127.0.0.1 -P 4000 -e "CREATE DATABASE IF NOT EXISTS $(MYSQL_DB_NAME);"

tidb-down: ## 🐯 Stop and remove the TiDB database container
	@echo "Stopping and removing TiDB container..."
	docker stop tidb-server || true
	docker rm tidb-server || true

tidb-logs: ## 🐯 View the logs of the TiDB container
	@echo "Following TiDB logs..."
	docker logs -f tidb-server

tidb-connect: ## 🐯 Connect to the TiDB database using MySQL Shell
	@echo "Connecting to TiDB database on port 4000 using mysqlsh..."
	mysqlsh --sql -u root -h 127.0.0.1 -P 4000
