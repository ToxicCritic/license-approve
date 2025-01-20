# Makefile

BINARY=LicenseApp
BUILD_DIR=build
CONFIG_DIR=config

build: 
	@echo "Building the application..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) cmd/main.go
	@echo "Copying configuration files..."
	cp -r $(CONFIG_DIR) $(BUILD_DIR)/

run: build
	@echo "Running the application..."
	$(BUILD_DIR)/$(BINARY)

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)