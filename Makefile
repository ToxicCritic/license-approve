# Makefile (located at the root of licence-approval)

# Binary names
CLIENT_BINARY=client
SERVER_BINARY=server
MOCK_OAUTH_BINARY=mock-oauth-server

# Source directories
CLIENT_DIR=client
SERVER_DIR=server
MOCK_OAUTH_DIR=mock-oauth-server/cmd

# Build directories
BUILD_DIR=build
CLIENT_BUILD_DIR=$(BUILD_DIR)/client
SERVER_BUILD_DIR=$(BUILD_DIR)/server
MOCK_OAUTH_BUILD_DIR=$(BUILD_DIR)/mock-oauth-server

# Go build flags
GO_FLAGS=-ldflags="-s -w"

# Declare phony targets
.PHONY: all build build-client build-server build-mock-oauth run run-client run-server run-mock-oauth clean help init-work

## ===== Default (help) =====
all: help

## ===== Help =====
help:
	@echo "================================================================"
	@echo "Multi-module Makefile for the licence-approval project"
	@echo ""
	@echo "Available targets:"
	@echo "  init-work         Initialize or update go.work with client, server, mock-oauth-server"
	@echo "  build             Build all modules"
	@echo "  build-client      Build the client module"
	@echo "  build-server      Build the server module"
	@echo "  build-mock-oauth  Build the mock OAuth2.0 server"
	@echo "  run               Build and run all modules (client, server, mock-oauth) concurrently"
	@echo "  run-client        Build and run only the client"
	@echo "  run-server        Build and run only the server"
	@echo "  run-mock-oauth    Build and run only the mock OAuth2.0 server"
	@echo "  clean             Remove all build artifacts"
	@echo "  help              Display this help message"
	@echo "================================================================"

## ===== init-work (go.work) =====
init-work:
	@echo "[WORK] Initializing go.work for client, server, mock-oauth..."
	go work init ./$(CLIENT_DIR) ./$(SERVER_DIR) ./$(MOCK_OAUTH_DIR)
	@echo "[WORK] go.work has been initialized/updated:"
	@cat go.work

## ===== Build all =====
build: build-client build-server build-mock-oauth

## ===== Build client =====
build-client:
	@echo "[CLIENT] Building client module..."
	mkdir -p $(CLIENT_BUILD_DIR)
	go build $(GO_FLAGS) -o $(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) ./$(CLIENT_DIR)
	@echo "[CLIENT] Copying config.json..."
	cp $(CLIENT_DIR)/config.json $(CLIENT_BUILD_DIR)/ || true
	@echo "[CLIENT] Done. Binary at $(CLIENT_BUILD_DIR)/$(CLIENT_BINARY)"

## ===== Build server =====
build-server:
	@echo "[SERVER] Building server module..."
	mkdir -p $(SERVER_BUILD_DIR)
	go build $(GO_FLAGS) -o $(SERVER_BUILD_DIR)/$(SERVER_BINARY) ./$(SERVER_DIR)

	@echo "[SERVER] Copying keys..."
	mkdir -p $(SERVER_BUILD_DIR)/config/keys
	cp $(SERVER_DIR)/config/keys/private_key.pem $(SERVER_BUILD_DIR)/config/keys/ || true
	cp $(SERVER_DIR)/config/keys/public_key.pem  $(SERVER_BUILD_DIR)/config/keys/ || true

	@echo "[SERVER] Done. Binary at $(SERVER_BUILD_DIR)/$(SERVER_BINARY)"

## ===== Build mock-oauth =====
build-mock-oauth:
	@echo "[MOCK-OAUTH] Building mock OAuth2.0 server..."
	mkdir -p $(MOCK_OAUTH_BUILD_DIR)
	go build $(GO_FLAGS) -o $(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY) ./$(MOCK_OAUTH_DIR)

	@echo "[MOCK-OAUTH] Copying config.json and certs..."
	mkdir -p $(MOCK_OAUTH_BUILD_DIR)/certs
	cp $(MOCK_OAUTH_DIR)/config.json $(MOCK_OAUTH_BUILD_DIR)/ || true
	cp $(MOCK_OAUTH_DIR)/certs/mock-oauth.crt $(MOCK_OAUTH_BUILD_DIR)/certs/ || true
	cp $(MOCK_OAUTH_DIR)/certs/mock-oauth.key $(MOCK_OAUTH_BUILD_DIR)/certs/ || true

	@echo "[MOCK-OAUTH] Done. Binary at $(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY)"

## ===== Run all =====
run: build
	@echo "[RUN] Starting all services in the background..."
	$(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY) &
	MOCK_OAUTH_PID=$!
	$(SERVER_BUILD_DIR)/$(SERVER_BINARY) &
	SERVER_PID=$!
	$(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) &
	CLIENT_PID=$!
	@echo "[RUN] All services launched. Press Ctrl+C to stop."
	wait $(MOCK_OAUTH_PID) $(SERVER_PID) $(CLIENT_PID)

## ===== Run client only =====
run-client: build-client
	@echo "[RUN-CLIENT] Launching client in this terminal..."
	cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY)

## ===== Run server only =====
run-server: build-server
	@echo "[RUN-SERVER] Launching server in this terminal..."
	cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY)

## ===== Run mock-oauth only =====
run-mock-oauth: build-mock-oauth
	@echo "[RUN-MOCK] Launching mock-oauth-server in this terminal..."
	cd $(MOCK_OAUTH_BUILD_DIR) && ./$(MOCK_OAUTH_BINARY)

## ===== Clean =====
clean:
	@echo "[CLEAN] Removing build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "[CLEAN] Done."