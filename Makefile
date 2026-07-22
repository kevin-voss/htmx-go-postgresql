.DEFAULT_GOAL := help

APP_NAME := forgeboard
COMPOSE := docker compose

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make dev       Start the complete development environment"
	@echo "  make stop      Stop containers"
	@echo "  make reset     Delete containers and database data"
	@echo "  make test      Run all tests"
	@echo "  make lint      Run Go checks"
	@echo "  make migrate   Apply database migrations"
	@echo "  make seed      Insert development data"
	@echo "  make logs      Follow application logs"

.PHONY: dev
dev:
	$(COMPOSE) up --build

.PHONY: stop
stop:
	$(COMPOSE) down

.PHONY: reset
reset:
	$(COMPOSE) down --volumes --remove-orphans

.PHONY: test
test:
	$(COMPOSE) run --rm app sh -c "go run ./cmd/migrate up && go test ./..."

.PHONY: lint
lint:
	$(COMPOSE) run --rm app sh -c \
		"go vet ./... && gofmt -l ."

.PHONY: migrate
migrate:
	$(COMPOSE) run --rm app \
		go run ./cmd/migrate up

.PHONY: seed
seed:
	$(COMPOSE) run --rm app \
		go run ./cmd/seed

.PHONY: logs
logs:
	$(COMPOSE) logs --follow app
