# Makefile

# Имена бинарных файлов
SERVER_BINARY=server
CLIENT_BINARY=client
MOCK_OAUTH_BINARY=mock-oauth-server

# Директории исходного кода
SERVER_DIR=server
CLIENT_DIR=client
MOCK_OAUTH_DIR=mock-oauth-server
COMMON_DIR=common

# Директории сборки
BUILD_DIR=build
SERVER_BUILD_DIR=$(BUILD_DIR)/server
CLIENT_BUILD_DIR=$(BUILD_DIR)/client
MOCK_OAUTH_BUILD_DIR=$(BUILD_DIR)/mock-oauth-server

# Флаги для сборки Go
GO_FLAGS=-ldflags="-s -w"

# Отключение эхо команд для чистого вывода
.SILENT:

# Объявление целей как phony для избежания конфликтов с файлами
.PHONY: all build build-server build-client build-mock-oauth run run-server run-client run-mock-oauth clean help

# Целевая по умолчанию: показать помощь
all: help

# Цель help: выводит доступные команды
help:
	@echo "Makefile для проекта LICENCE-APPROVAL"
	@echo ""
	@echo "Доступные цели:"
	@echo "  make build                  Сборка всех сервисов (server, client, mock-oauth-server)"
	@echo "  make run                    Сборка и запуск всех сервисов"
	@echo "  make build-server           Сборка сервера"
	@echo "  make run-server             Сборка и запуск сервера"
	@echo "  make build-client           Сборка клиента"
	@echo "  make run-client             Сборка и запуск клиента"
	@echo "  make build-mock-oauth       Сборка mock OAuth2.0 сервера аутентификации"
	@echo "  make run-mock-oauth         Сборка и запуск mock OAuth2.0 сервера аутентификации"
	@echo "  make clean                  Очистка артефактов сборки"
	@echo "  make help                   Показать это сообщение"
	@echo ""

# Цель build: сборка всех сервисов
build: build-server build-client build-mock-oauth

# Цель build-server: сборка сервера и копирование конфигурационных файлов
build-server:
	@echo "Сборка сервера..."
	mkdir -p $(SERVER_BUILD_DIR)
	cd $(SERVER_DIR) && go build $(GO_FLAGS) -o ../../$(SERVER_BUILD_DIR)/$(SERVER_BINARY) .
	@echo "Копирование конфигурационных файлов в build/server..."
	cp -r $(SERVER_DIR)/config $(SERVER_BUILD_DIR)/ || true

# Цель build-client: сборка клиента
build-client:
	@echo "Сборка клиента..."
	mkdir -p $(CLIENT_BUILD_DIR)
	cd $(CLIENT_DIR) && go build $(GO_FLAGS) -o ../../$(CLIENT_BUILD_DIR)/$(CLIENT_BINARY) .

# Цель build-mock-oauth: сборка mock OAuth2.0 сервера аутентификации
build-mock-oauth:
	@echo "Сборка mock OAuth2.0 сервера..."
	mkdir -p $(MOCK_OAUTH_BUILD_DIR)
	cd $(MOCK_OAUTH_DIR) && go build $(GO_FLAGS) -o ../../$(MOCK_OAUTH_BUILD_DIR)/$(MOCK_OAUTH_BINARY) .

# Цель run: сборка и запуск всех сервисов с управлением процессами
run: build
	@echo "Запуск всех сервисов..."
	@bash -c '\
		# Функция для завершения всех процессов при выходе\
		function cleanup() { \
			echo "Завершение всех сервисов..."; \
			kill $$MOCK_OAUTH_PID $$SERVER_PID $$CLIENT_PID 2>/dev/null; \
			exit 0; \
		}; \
		trap cleanup SIGINT SIGTERM; \
		\
		# Запуск mock OAuth2.0 сервера в фоновом режиме\
		(cd $(MOCK_OAUTH_BUILD_DIR) && ./$(MOCK_OAUTH_BINARY)) & \
		MOCK_OAUTH_PID=$$!; \
		echo "Mock OAuth2.0 сервер запущен с PID $$MOCK_OAUTH_PID"; \
		\
		# Запуск серверного приложения в фоновом режиме\
		(cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY)) & \
		SERVER_PID=$$!; \
		echo "Сервер запущен с PID $$SERVER_PID"; \
		\
		# Запуск клиентского приложения в фоновом режиме\
		(cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY)) & \
		CLIENT_PID=$$!; \
		echo "Клиент запущен с PID $$CLIENT_PID"; \
		\
		# Ожидание завершения всех процессов\
		wait $$MOCK_OAUTH_PID $$SERVER_PID $$CLIENT_PID \
	'

# Цель run-server: сборка и запуск только сервера
run-server: build-server
	@echo "Запуск сервера..."
	@bash -c '\
		function cleanup() { \
			echo "Завершение сервера..."; \
			kill $$SERVER_PID 2>/dev/null; \
			exit 0; \
		}; \
		trap cleanup SIGINT SIGTERM; \
		\
		(cd $(SERVER_BUILD_DIR) && ./$(SERVER_BINARY)) & \
		SERVER_PID=$$!; \
		echo "Сервер запущен с PID $$SERVER_PID"; \
		\
		wait $$SERVER_PID \
	'

# Цель run-client: сборка и запуск только клиента
run-client: build-client
	@echo "Запуск клиента..."
	@bash -c '\
		function cleanup() { \
			echo "Завершение клиента..."; \
			kill $$CLIENT_PID 2>/dev/null; \
			exit 0; \
		}; \
		trap cleanup SIGINT SIGTERM; \
		\
		(cd $(CLIENT_BUILD_DIR) && ./$(CLIENT_BINARY)) & \
		CLIENT_PID=$$!; \
		echo "Клиент запущен с PID $$CLIENT_PID"; \
		\
		wait $$CLIENT_PID \
	'

# Цель run-mock-oauth: сборка и запуск только mock OAuth2.0 сервера
run-mock-oauth: build-mock-oauth
	@echo "Запуск mock OAuth2.0 сервера..."
	@bash -c '\
		function cleanup() { \
			echo "Завершение mock OAuth2.0 сервера..."; \
			kill $$MOCK_OAUTH_PID 2>/dev/null; \
			exit 0; \
		}; \
		trap cleanup SIGINT SIGTERM; \
		\
		(cd $(MOCK_OAUTH_BUILD_DIR) && ./$(MOCK_OAUTH_BINARY)) & \
		MOCK_OAUTH_PID=$$!; \
		echo "Mock OAuth2.0 сервер запущен с PID $$MOCK_OAUTH_PID"; \
		\
		wait $$MOCK_OAUTH_PID \
	'

# Цель clean: очистка артефактов сборки
clean:
	@echo "Очистка артефактов сборки..."
	rm -rf $(BUILD_DIR)
	@echo "Очистка завершена."