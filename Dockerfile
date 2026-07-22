# Development image: compile/run Go inside the container.
FROM golang:1.26-alpine AS development

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

# Migrate then run the web app (see scripts/dev-entrypoint.sh).
# Invoked via sh so the bind-mounted script need not be executable.
CMD ["sh", "./scripts/dev-entrypoint.sh"]

# ---------------------------------------------------------------------------
# Production multi-stage build
# ---------------------------------------------------------------------------

FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/forgeboard ./cmd/web \
	&& CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22 AS production

RUN apk add --no-cache ca-certificates \
	&& adduser -D -H -u 10001 forgeboard

WORKDIR /app

COPY --from=builder /out/forgeboard /usr/local/bin/forgeboard
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY --from=builder /src/db/migrations /app/db/migrations

USER forgeboard

EXPOSE 8080

ENV APP_ENV=production \
	APP_ADDRESS=:8080

# Templates and static assets are embedded in the binary (web/embed.go).
# Run migrations separately from /app: migrate up
CMD ["forgeboard"]
