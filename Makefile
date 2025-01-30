# Makefile (расположен в корне проекта licence-approval)

# Binary names
CLIENT_BINARY=client
SERVER_BINARY=server
MOCK_OAUTH_BINARY=mock-oauth-server

# Source directories
CLIENT_DIR=client
SERVER_DIR=server
MOCK_OAUTH_DIR=mock-oauth-server

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
	@echo "============================================================"
	@echo "Makefile for multi-module Go project (client, server, mock-oauth-server)"
	@echo ""
	@echo "Targets:"
	@echo "  init-work         Initialize or update go.work (workspace)"
	@echo "  build             Build all modules"
	@echo "  build-client      Build the client module"
	@echo "  build-server      Build the server module"
	@echo "  build-mock-oauth  Build the mock OAuth2.0 server module"
	@echo "  run               Build and run all modules concurrently"
	@echo "  run-client        Build and run only client"
	@echo "  run-server        Build and run only server"
	@echo "  run-mock-oauth    Build and run only mock-oauth-server"
	@echo "  clean             Remove build artifacts"
	@echo "  help              Show this help message"
	@echo "============================================================"

## ===== init-work (go.work) =====
init-work:
	@echo "[WORK] Initializing go.work with ./client, ./server, ./mock-oauth-server..."
	go work init ./$(CLIENT_DIR) ./$(SERVER_DIR) ./$(MOCK_OAUTH_DIR)
	@echo "go.work has been (re)initialized:"
	@cat go.work

## ===== Build all =====
build: build-client build-server build-mock-oauth

## ===== Build client =====
build-client:
	@echo "[CLIENT] Building the client module..."
	mkdir -p $(CLIENT_BUILD_DIR)
	go build $(GO_FLAGS) -o $(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) ./$(CLIENT_DIR)
	@echo "[CLIENT] Copying config.json..."
	cp $(CLIENT_DIR)/config.json $(CLIENT_BUILD_DIR)/ || true
	@echo "[CLIENT] Built client -> $(CLIENT_BUILD_DIR)/$(CLIENT_BINARY)"

## ===== Build server =====
build-server:
	@echo "[SERVER] Building the server module..."
	mkdir -p $(SERVER_BUILD_DIR)
	go build $(GO_FLAGS) -o $(SERVER_BUILD_DIR)/$(SERVER_BINARY) ./$(SERVER_DIR)

	@echo "[SERVER] Copying keys..."
	mkdir -p $(SERVER_BUILD_DIR)/config/keys
	cp server/config/keys/private_key.pem $(SERVER_BUILD_DIR)/config/keys/ || true
	cp server/config/keys/public_key.pem  $(SERVER_BUILD_DIR)/config/keys/ || true

	@echo "[SERVER] Built server -> $(SERVER_BUILD_DIR)/$(SERVER_BINARY)"

## ===== Build mock-oauth-server =====
build-mock-oauth:
	@echo "[OAUTH] Building the mock OAuth2.0 server..."
	mkdir -p $(MOCK_OAUTH_BUILD_DIR)
	go build $(GO_FLAGS) -o $(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY) ./$(MOCK_OAUTH_DIR)

	@echo "[OAUTH] Copying config.json and certs..."
	mkdir -p $(MOCK_OAUTH_BUILD_DIR)/certs
	cp $(MOCK_OAUTH_DIR)/config.json $(MOCK_OAUTH_BUILD_DIR)/ || true
	cp $(MOCK_OAUTH_DIR)/certs/mock-oauth.crt $(MOCK_OAUTH_BUILD_DIR)/certs/ || true
	cp $(MOCK_OAUTH_DIR)/certs/mock-oauth.key $(MOCK_OAUTH_BUILD_DIR)/certs/ || true

	@echo "[OAUTH] Built mock-oauth-server -> $(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY)"

## ===== Run all =====
run: build
	@echo "[RUN] Launching all services concurrently..."
	$(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY) &
	MOCK_OAUTH_PID=$!
	$(SERVER_BUILD_DIR)/$(SERVER_BINARY) &
	SERVER_PID=$!
	$(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) &
	CLIENT_PID=$!
	@echo "[RUN] All services launched in background. Press Ctrl+C to stop."
	wait $(MOCK_OAUTH_PID) $(SERVER_PID) $(CLIENT_PID)

## ===== Run client only =====
run-client: build-client
	@echo "[RUN-CLIENT] Launching client in current terminal..."
	cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY)

## ===== Run server only =====
run-server: build-server
	@echo "[RUN-SERVER] Launching server in current terminal..."
	cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY)

## ===== Run mock-oauth-server only =====
run-mock-oauth: build-mock-oauth
	@echo "[RUN-OAUTH] Launching mock-oauth-server in current terminal..."
	cd $(MOCK_OAUTH_BUILD_DIR) && ./$(MOCK_OAUTH_BINARY)

## ===== Clean =====
clean:
	@echo "[CLEAN] Removing build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "[CLEAN] Done."