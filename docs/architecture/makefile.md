# Makefile specification

Primary command:

```bash
make dev
```

It should:

1. build the development image
2. start PostgreSQL
3. start Mailpit
4. run database migrations (via app entrypoint)
5. start the application
6. show application logs

## Suggested Makefile

```makefile
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
	$(COMPOSE) run --rm app go test ./...

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
```

## Happy path

```bash
git clone ...
cd forgeboard
make dev
```

Then open `http://localhost:8080`.

## Related

- Docker: [docker.md](docker.md)
- Definition of done: [../DEFINITION_OF_DONE.md](../DEFINITION_OF_DONE.md)
