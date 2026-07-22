#!/bin/sh
set -e

# Development entrypoint: migrate, then start the web app.
# Used by the development Docker image so `make dev` is enough.
go run ./cmd/migrate up
exec go run ./cmd/web
