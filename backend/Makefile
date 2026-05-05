# ============================================
# Personal Blog Backend - Makefile
# ============================================

APP_NAME := blog-server
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/server/main.go
MIGRATE_DIR := ./migrations
DATABASE_URL ?= postgres://blog_user:blog_password@localhost:5432/blog_dev?sslmode=disable

.PHONY: build run test migrate migrate-down migrate-create docker-build docker-run clean lint fmt help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## build: Build the Go binary
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Binary built at $(BUILD_DIR)/$(APP_NAME)"

## run: Run the server locally (requires .env or env vars)
run:
	@echo "Starting server..."
	go run $(MAIN_PATH)

## test: Run all tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

## test-short: Run tests without verbose output
test-short:
	go test -race -cover ./...

## migrate: Run database migrations up
migrate:
	@echo "Running migrations..."
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" up

## migrate-down: Rollback the last migration
migrate-down:
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATE_DIR) -database "$(DATABASE_URL)" down 1

## migrate-create: Create a new migration file (usage: make migrate-create name=create_users)
migrate-create:
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir $(MIGRATE_DIR) -seq $(name)

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

## docker-run: Run the Docker image
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 --env-file ../.env $(APP_NAME):latest

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format Go source files
fmt:
	gofmt -s -w .

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."
