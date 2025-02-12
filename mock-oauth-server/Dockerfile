# =====================
#  Стадия сборки (Go)
# =====================
FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Копируем всё содержимое mock-oauth-server (cmd, certs, и т.д.)
COPY . ./

# Собираем бинарник
RUN mkdir -p build/mock-oauth
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/mock-oauth/mock-oauth-server ./cmd/mockserver.go

# =====================
#  Стадия "release"
# =====================
FROM alpine:3.18 AS final
WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/build/mock-oauth/mock-oauth-server ./mock-oauth-server

# Копируем сертификаты (если нужны)
COPY certs/ certs/

EXPOSE 8081
CMD ["./mock-oauth-server"]