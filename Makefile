SHELL := /bin/bash

# Override these at invocation time if your commands differ:
#   make ui UI_CMD="npx vite"
#   make server SERVER_CMD="go run ./cmd/server"
UI_CMD ?= npx tsx client.ts
SERVER_CMD ?= go run ./cmd/server

.PHONY: help ui server dev test test-ts test-go

help:
	@echo "Available targets:"
	@echo "  make ui       - Run frontend/UI process (TypeScript)"
	@echo "  make server   - Run backend/server process (Go)"
	@echo "  make test     - Run both TypeScript and Go tests"
	@echo "  make test-ts  - Run TypeScript tests"
	@echo "  make test-go  - Run Go tests"
	@echo "  make dev      - Print commands to run both processes"
	@echo ""
	@echo "Configurable variables:"
	@echo "  UI_CMD        (default: $(UI_CMD))"
	@echo "  SERVER_CMD    (default: $(SERVER_CMD))"

ui:
	@echo "Starting UI with: $(UI_CMD)"
	@$(UI_CMD)

server:
	@echo "Starting server with: $(SERVER_CMD)"
	@$(SERVER_CMD)

test: test-ts test-go

test-ts:
	@npx vitest run

test-go:
	@go test ./...

dev:
	@echo "Run these in separate terminals:"
	@echo "  make ui UI_CMD=\"$(UI_CMD)\""
	@echo "  make server SERVER_CMD=\"$(SERVER_CMD)\""
