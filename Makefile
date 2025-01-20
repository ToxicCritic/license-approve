# Makefile

# Имена бинарников
SERVER_BINARY=server
CLIENT_BINARY=client

# Директория для сборки
BUILD_DIR=build

# Директория с конфигурационными файлами
CONFIG_DIR=config

# Цель build-server: сборка серверного сервиса и копирование конфигурационных файлов
build-server:
	@echo "Building the server application..."
	mkdir -p $(BUILD_DIR)/server
	go build -o $(BUILD_DIR)/server/$(SERVER_BINARY) server/main.go
	@echo "Copying configuration files to build/server..."
	cp -r $(CONFIG_DIR) $(BUILD_DIR)/server/

# Цель run-server: сборка и запуск серверного сервиса
run-server: build-server
	@echo "Running the server application..."
	cd $(BUILD_DIR)/server && ./$(SERVER_BINARY)

# Цель build-client: сборка клиентского сервиса
build-client:
	@echo "Building the client application..."
	mkdir -p $(BUILD_DIR)/client
	go build -o $(BUILD_DIR)/client/$(CLIENT_BINARY) client/main.go

# Цель run-client: сборка и запуск клиентского сервиса
run-client: build-client
	@echo "Running the client application..."
	cd $(BUILD_DIR)/client && ./$(CLIENT_BINARY)

# Цель build: сборка всех сервисов
build: build-server build-client

# Цель run: сборка и запуск всех сервисов
run: build
	@echo "Starting server and client..."
	( cd $(BUILD_DIR)/server && ./$(SERVER_BINARY) ) &
	( cd $(BUILD_DIR)/client && ./$(CLIENT_BINARY) ) &

# Цель clean: удаление артефактов сборки
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)