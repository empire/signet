SHELL := /bin/bash

UI_CMD ?= npx vite --host
SERVER_CMD ?= go run ./cmd/server

.PHONY: help ui server dev test test-ts test-go

help:
	@echo "Available targets:"
	@echo "  make ui       - Run frontend/UI process"
	@echo "  make server   - Run backend/server process"
	@echo "  make test     - Run both TypeScript and Go tests"
	@echo "  make test-ts  - Run TypeScript tests"
	@echo "  make test-go  - Run Go tests"
	@echo "  make dev      - Print commands to run both processes"

ui:
	@$(UI_CMD)

server:
	@$(SERVER_CMD)

test: test-ts test-go

test-ts:
	@npx vitest run

test-go:
	@go test ./...

dev:
	@echo "Run these in separate terminals:"
	@echo "  make ui"
	@echo "  make server"
