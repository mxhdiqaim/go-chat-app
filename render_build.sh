#!/usr/bin/env bash
# exit on error
set -o errexit

# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Build the main application
go build -o app ./cmd/api