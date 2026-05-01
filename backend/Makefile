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
	@go run ./cmd/api -c config/config.yaml

run-bot:
	@go run ./cmd/bot -c config/config.yaml

test:
	@go test -v -race -coverprofile=coverage.out ./...

lint:
	@PATH="$(shell go env GOPATH)/bin:$$PATH" golangci-lint run

migrate-up:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down-to 0

clean:
	@rm -rf bin/
	@rm -f coverage.out

