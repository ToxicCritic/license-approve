FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN mkdir -p build/mock-oauth
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/mock-oauth/mock-oauth-server ./cmd/mockserver.go

FROM alpine:latest AS final
WORKDIR /app

COPY --from=builder /app/build/mock-oauth/mock-oauth-server ./mock-oauth-server

COPY certs/ certs/

EXPOSE 8081
CMD ["./mock-oauth-server"]