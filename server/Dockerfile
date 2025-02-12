# =========================================
#  Stage 1: Сборка Go-бинарника
# =========================================
FROM golang:1.23 AS builder
WORKDIR /app


# 1) Скопируем go.mod/go.sum из папки server/
COPY go.mod go.sum ./
RUN go mod download

# 2) Копируем код (main.go, config, etc.) — всё, что лежит в server/
COPY . ./

# 3) Собираем бинарник
RUN mkdir -p build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/server ./main.go


# =========================================
#  Stage 2: Финальный образ
# =========================================
FROM alpine:3.18 AS final
WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/build/server ./server

# Копируем .env (если нужно)
COPY .env .env

# Копируем config/ (где keys, certs и т.д.)
COPY config config

# Сервер слушает HTTPS на 8443
EXPOSE 8443

# Команда запуска: указываем ИМЯ бинарника (./server), а не ./
CMD ["./server"]