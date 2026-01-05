.PHONY: help build run-api run-bot test lint migrate-up migrate-down clean

help:
	@echo "Available commands:"
	@echo "  make build       - Build all binaries"
	@echo "  make run-api     - Run API server"
	@echo "  make run-bot     - Run Telegram bot"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run linter"
	@echo "  make migrate-up  - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make clean       - Clean build artifacts"

build:
	@go build -o bin/api ./cmd/api
	@go build -o bin/bot ./cmd/bot

run-api:
	@go run ./cmd/api

run-bot:
	@go run ./cmd/bot

test:
	@go test -v -race -coverprofile=coverage.out ./...

lint:
	@golangci-lint run

migrate-up:
	@./scripts/migrate.sh

migrate-down:
	@./scripts/rollback.sh

clean:
	@rm -rf bin/
	@rm -f coverage.out

