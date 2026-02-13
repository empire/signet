SHELL := /bin/bash

# Override these at invocation time if your commands differ:
#   make ui UI_CMD="pnpm --dir web dev"
#   make server SERVER_CMD="go run ./cmd/api"
UI_CMD ?= npm run dev
SERVER_CMD ?= go run ./server

.PHONY: help ui server dev

help:
	@echo "Available targets:"
	@echo "  make ui      - Run frontend/UI process"
	@echo "  make server  - Run backend/server process"
	@echo "  make dev     - Print commands to run both processes"
	@echo ""
	@echo "Configurable variables:"
	@echo "  UI_CMD       (default: $(UI_CMD))"
	@echo "  SERVER_CMD   (default: $(SERVER_CMD))"

ui:
	@echo "Starting UI with: $(UI_CMD)"
	@$(UI_CMD)

server:
	@echo "Starting server with: $(SERVER_CMD)"
	@$(SERVER_CMD)

dev:
	@echo "Run these in separate terminals:"
	@echo "  make ui UI_CMD=\"$(UI_CMD)\""
	@echo "  make server SERVER_CMD=\"$(SERVER_CMD)\""
