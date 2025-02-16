FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN mkdir -p build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/client ./main.go


FROM alpine:latest AS final
WORKDIR /app

COPY --from=builder /app/build/client ./client

COPY config.json config.json

RUN mkdir -p server/config/certs
COPY license-srv-srt/server.crt server/config/certs/server.crt

CMD ["./client"]