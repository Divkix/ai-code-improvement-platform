# ==============================================================================
# Makefile for AI Code Improvement Platform
#
# Environments:
#  - prod (default): For production-like deployments.
#  - dev: For local development with hot-reloading.
#
# Usage:
#  - make <command>          (runs command in 'prod' environment)
#  - make <command> env=dev  (runs command in 'dev' environment)
# ==============================================================================

# Phony targets prevent conflicts with files of the same name.
.PHONY: help up down clean logs ps sh generate build dev backend-dev

# Default environment is 'prod'
env ?= prod

# Set compose file based on the environment
COMPOSE_FILE := docker-compose.yml
ifeq ($(env), dev)
	COMPOSE_FILE := docker-compose.dev.yml
endif

# Use a consistent docker compose command
DOCKER_COMPOSE := docker compose -f $(COMPOSE_FILE)

# Default target executed when 'make' is run without arguments.
default: help

# ------------------------------------------------------------------------------
# Main Commands
# ------------------------------------------------------------------------------

## up: Build and start all services in the background.
up:
	@echo "🚀 Starting services for $(env) environment..."
	@$(DOCKER_COMPOSE) up --build -d

## down: Stop all running services.
down:
	@echo "🛑 Stopping services for $(env) environment..."
	@$(DOCKER_COMPOSE) down

## clean: Stop all services and remove all associated data volumes.
clean:
	@echo "🧼 Cleaning up services and data for $(env) environment..."
	@$(DOCKER_COMPOSE) down --volumes --remove-orphans

# ------------------------------------------------------------------------------
# Utility Commands
# ------------------------------------------------------------------------------

## logs: View logs from all services. Use 'service=<name>' to specify one.
logs:
	@echo "📜 Tailing logs for $(env) environment..."
	@$(DOCKER_COMPOSE) logs -f $(service)

## ps: List all running containers for the environment.
ps:
	@echo "📊 Displaying status for $(env) environment..."
	@$(DOCKER_COMPOSE) ps

## sh: Get a shell inside a running service (default: backend).
sh:
	@service_name=$$(or $(service),backend); \
	echo "💻 Accessing shell in '$$service_name' container..."; \
	@$(DOCKER_COMPOSE) exec $$service_name sh

# ------------------------------------------------------------------------------
# Development Commands
# ------------------------------------------------------------------------------

## generate: Generate Go code from OpenAPI specification
generate:
	@echo "🔧 Generating Go code from OpenAPI spec..."
	@mkdir -p backend/internal/generated
	@cd backend && oapi-codegen -generate types -o internal/generated/types.go -package generated api/openapi.yaml
	@cd backend && oapi-codegen -generate gin -o internal/generated/server.go -package generated api/openapi.yaml
	@cd backend && oapi-codegen -generate client -o internal/generated/client.go -package generated api/openapi.yaml
	@echo "✅ Code generation complete"

## build: Generate code and build the backend binary
build: generate
	@echo "🔨 Building backend binary..."
	@cd backend && go build -o bin/server cmd/server/main.go
	@echo "✅ Build complete"

## dev: Run backend in development mode (with code generation)
backend-dev: generate
	@echo "🚀 Starting backend in development mode..."
	@cd backend && go run cmd/server/main.go

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------

## help: Display this help message.
help:
	@echo "Makefile Commands:"
	@echo ""
	@echo "Usage: make <command> [env=prod|dev] [service=<name>]"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
