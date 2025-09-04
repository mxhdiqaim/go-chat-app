#!/usr/bin/env bash
# exit on error
set -o errexit

# Install deps
go mod tidy
go install github.com/pressly/goose/v3/cmd/goose@latest

goose -dir ./migrations postgres "$DATABASE_URL" up

# Build the main application
go build -o app ./cmd/api