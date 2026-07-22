# Docker development environment

## Services

```text
app
database
mailpit
```

## Target `compose.yaml`

```yaml
services:
  app:
    build:
      context: .
      target: development
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      APP_ADDRESS: ":8080"
      DATABASE_URL: postgres://forgeboard:forgeboard@database:5432/forgeboard?sslmode=disable
      SMTP_HOST: mailpit
      SMTP_PORT: "1025"
    volumes:
      - .:/app
    depends_on:
      database:
        condition: service_healthy

  database:
    image: postgres:18-alpine
    environment:
      POSTGRES_DB: forgeboard
      POSTGRES_USER: forgeboard
      POSTGRES_PASSWORD: forgeboard
    ports:
      - "5432:5432"
    volumes:
      - forgeboard_database:/var/lib/postgresql/data
    healthcheck:
      test:
        - CMD-SHELL
        - pg_isready -U forgeboard -d forgeboard
      interval: 5s
      timeout: 5s
      retries: 10

  mailpit:
    image: axllent/mailpit
    ports:
      - "8025:8025"
      - "1025:1025"

volumes:
  forgeboard_database:
```

## Local URLs

```text
Application: http://localhost:8080
Mailpit:     http://localhost:8025
PostgreSQL:  localhost:5432
```

## Entrypoint expectation

Development entrypoint should run migrations then the app so `make dev` alone is enough:

```sh
go run ./cmd/migrate up
go run ./cmd/web
```

## Related

- Makefile: [makefile.md](makefile.md)
- Implementation steps for Docker / Makefile under [../implementation/steps/](../implementation/steps/)
