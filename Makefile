SHELL := /bin/bash
BINARY := bin/api
MIGRATE_PATH := migrations
DB_DSN ?= mysql://gdc:gdc_secret@tcp(localhost:3306)/gdc

.PHONY: help run build tidy test test-race test-integration cover lint vet fmt swag migrate-up migrate-down migrate-create docker-up docker-down logs clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Run the API locally
	go run ./cmd/api

build: ## Build the API binary into bin/
	go build -trimpath -o $(BINARY) ./cmd/api

tidy: ## Sync go.mod / go.sum
	go mod tidy

fmt: ## Format and fix imports
	go fmt ./...
	go run golang.org/x/tools/cmd/goimports@latest -w .

vet: ## Run go vet
	go vet ./...

test: ## Run unit tests (no external services required)
	go test ./...

test-race: ## Run unit tests with the race detector
	go test -race -count=1 ./...

test-integration: ## Run integration tests (requires `make docker-up && make migrate-up`)
	go test -tags integration -count=1 ./test/integration/...

cover: ## Produce a coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

lint: ## Run golangci-lint
	golangci-lint run ./...

swag: ## Regenerate the OpenAPI/Swagger docs
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

migrate-up: ## Apply all migrations
	migrate -path $(MIGRATE_PATH) -database "$(DB_DSN)" up

migrate-down: ## Roll back the last migration
	migrate -path $(MIGRATE_PATH) -database "$(DB_DSN)" down 1

migrate-create: ## Create a migration: make migrate-create name=add_table
	migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(name)

docker-up: ## Start MySQL, Redis and apply migrations
	docker compose up -d --build

docker-down: ## Stop all services
	docker compose down

logs: ## Tail API logs
	docker compose logs -f api

clean: ## Remove build artifacts
	rm -rf bin coverage.out coverage.html
seed: ## Seed dummy users and tasks (requires DB running)
	go run ./cmd/seed
