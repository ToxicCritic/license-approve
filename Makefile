# Makefile

# Binary names
SERVER_BINARY=server
CLIENT_BINARY=client
MOCK_OAUTH_BINARY=mock-oauth-server

# Source directories
SERVER_DIR=server
CLIENT_DIR=client
MOCK_OAUTH_DIR=mock-oauth-server
COMMON_DIR=common

# Build directories
BUILD_DIR=build
SERVER_BUILD_DIR=$(BUILD_DIR)/server
CLIENT_BUILD_DIR=$(BUILD_DIR)/client
MOCK_OAUTH_BUILD_DIR=$(BUILD_DIR)/mock-oauth-server

# Go build flags (e.g., for optimizations)
GO_FLAGS=-ldflags="-s -w"

# Suppress command echoing for cleaner output
.SILENT:

# Declare phony targets to avoid conflicts with files of the same name
.PHONY: all build build-server build-client build-mock-oauth run run-server run-client run-mock-oauth clean help

# Default target: display help
all: help

# Help target: displays available Makefile commands
help:
	@echo "Makefile for LICENCE-APPROVAL Project"
	@echo ""
	@echo "Available targets:"
	@echo "  make build            Build all services (server, client, mock-oauth-server)"
	@echo "  make run              Build and run all services"
	@echo "  make build-server     Build the server service"
	@echo "  make run-server       Build and run the server service"
	@echo "  make build-client     Build the client service"
	@echo "  make run-client       Build and run the client service"
	@echo "  make build-mock-oauth Build the mock-oauth-server service"
	@echo "  make run-mock-oauth   Build and run the mock-oauth-server service"
	@echo "  make clean            Clean build artifacts"
	@echo "  make help             Display this help message"
	@echo ""

# Build target: builds all services
build: build-server build-client build-mock-oauth

# Build-server target: builds the server service and copies configuration files
build-server:
	@echo "Building the server service..."
	mkdir -p $(SERVER_BUILD_DIR)
	cd $(SERVER_DIR) && go build $(GO_FLAGS) -o ../../$(SERVER_BUILD_DIR)/$(SERVER_BINARY) main.go
	@echo "Copying configuration files to build/server..."
	cp -r $(SERVER_DIR)/config $(SERVER_BUILD_DIR)/

# Build-client target: builds the client service
build-client:
	@echo "Building the client service..."
	mkdir -p $(CLIENT_BUILD_DIR)
	cd $(CLIENT_DIR) && go build $(GO_FLAGS) -o ../../$(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) main.go

# Build-mock-oauth target: builds the mock-oauth-server service
build-mock-oauth:
	@echo "Building the mock-oauth-server service..."
	mkdir -p $(MOCK_OAUTH_BUILD_DIR)
	cd $(MOCK_OAUTH_DIR) && go build $(GO_FLAGS) -o ../../$(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY) main.go

# Run target: builds and runs all services
run: build
	@echo "Starting all services..."
	# Start server in the background
	( cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY) ) &
	SERVER_PID=$!
	@echo "Server started with PID $$SERVER_PID"

	# Start mock-oauth-server in the background
	( cd $(MOCK_OAUTH_BUILD_DIR) && ./$(MOCK_OAUTH_BINARY) ) &
	MOCK_OAUTH_PID=$!
	@echo "Mock OAuth Server started with PID $$MOCK_OAUTH_PID"

	# Wait briefly to ensure the server starts before the client attempts to connect
	sleep 2

	# Start client in the background
	( cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY) ) &
	CLIENT_PID=$!
	@echo "Client started with PID $$CLIENT_PID"

	# Wait for all services to finish
	wait $$SERVER_PID $$MOCK_OAUTH_PID $$CLIENT_PID

# Run-server target: builds and runs only the server service
run-server: build-server
	@echo "Starting the server service..."
	cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY)

# Run-client target: builds and runs only the client service
run-client: build-client
	@echo "Starting the client service..."
	cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY)

# Run-mock-oauth target: builds and runs only the mock-oauth-server service
run-mock-oauth: build-mock-oauth
	@echo "Starting the mock-oauth-server service..."
	cd $(MOCK_OAUTH_BUILD_DIR) && ./$(MOCK_OAUTH_BINARY)

# Clean target: removes build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean completed."