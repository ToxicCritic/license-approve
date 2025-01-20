# Makefile

# Binary names
SERVER_BINARY=server
CLIENT_BINARY=client

# Source directories
SERVER_DIR=server
CLIENT_DIR=client

# Build directories
BUILD_DIR=build
SERVER_BUILD_DIR=$(BUILD_DIR)/server
CLIENT_BUILD_DIR=$(BUILD_DIR)/client

# Configuration directory (updated path)
CONFIG_DIR=server/config

# Go build flags (e.g., for optimizations)
GO_FLAGS=-ldflags="-s -w"

# Suppress command echoing for cleaner output
.SILENT:

# Declare phony targets to avoid conflicts with files of the same name
.PHONY: all build build-server build-client run run-server run-client clean help

# Default target: display help
all: help

# Help target: displays available Makefile commands
help:
	@echo "Makefile for LicenseApp Project"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        Build both server and client services"
	@echo "  make run          Build and run both server and client services"
	@echo "  make build-server Build the server service"
	@echo "  make run-server    Build and run the server service"
	@echo "  make build-client Build the client service"
	@echo "  make run-client    Build and run the client service"
	@echo "  make clean        Clean build artifacts"
	@echo "  make help         Display this help message"
	@echo ""

# Build target: builds both server and client
build: build-server build-client

# Build-server target: builds the server service and copies configuration files
build-server:
	@echo "Building the server service..."
	mkdir -p $(SERVER_BUILD_DIR)
	go build $(GO_FLAGS) -o $(SERVER_BUILD_DIR)/$(SERVER_BINARY) $(SERVER_DIR)/main.go
	@echo "Copying configuration files to build/server..."
	cp -r $(CONFIG_DIR) $(SERVER_BUILD_DIR)/

# Build-client target: builds the client service
build-client:
	@echo "Building the client service..."
	mkdir -p $(CLIENT_BUILD_DIR)
	go build $(GO_FLAGS) -o $(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) $(CLIENT_DIR)/main.go

# Run target: builds and runs both server and client services
run: build
	@echo "Starting server and client services..."
	# Start server in the background
	( cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY) ) &
	SERVER_PID=$!
	@echo "Server started with PID $$SERVER_PID"

	# Wait briefly to ensure the server starts before the client attempts to connect
	sleep 2

	# Start client in the background
	( cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY) ) &
	CLIENT_PID=$!
	@echo "Client started with PID $$CLIENT_PID"

	# Wait for both server and client to finish
	wait $$SERVER_PID $$CLIENT_PID

# Run-server target: builds and runs only the server service
run-server: build-server
	@echo "Starting the server service..."
	# Start server in the foreground
	( cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY) )
	# If you prefer to run the server in the background, uncomment the lines below:
	# ( cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY) ) &
	# SERVER_PID=$!
	# @echo "Server started with PID $$SERVER_PID"
	# wait $$SERVER_PID

# Run-client target: builds and runs only the client service
run-client: build-client
	@echo "Starting the client service..."
	# Start client in the foreground
	( cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY) )
	# If you prefer to run the client in the background, uncomment the lines below:
	# ( cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY) ) &
	# CLIENT_PID=$!
	# @echo "Client started with PID $$CLIENT_PID"
	# wait $$CLIENT_PID

# Clean target: removes build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean completed."