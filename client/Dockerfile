# =========================================
#  Stage 1: Сборка Go-бинарника
# =========================================
FROM golang:1.23 AS builder
WORKDIR /app

# Скопируем go.mod/go.sum (если они лежат в папке client/)
COPY go.mod go.sum ./
RUN go mod download

# Копируем код (main.go, pkg/, config.json, ...)
COPY . ./

# Собираем бинарник
RUN mkdir -p build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/client ./main.go

# =========================================
#  Stage 2: Финальный образ
# =========================================
FROM alpine:latest AS final
WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/build/client ./client

# Копируем config.json
COPY config.json config.json

RUN mkdir -p server/config/certs
# Копируем серверный сертификат
COPY license-srv-srt/server.crt server/config/certs/server.crt

# Клиент не слушает входящего порта, поэтому EXPOSE не нужен
CMD ["./client"]