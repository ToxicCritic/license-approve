FROM golang:1.23 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN mkdir -p build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/server ./main.go


FROM alpine:latest AS final
WORKDIR /app

COPY --from=builder /app/build/server ./server

COPY .env .env

COPY config config

EXPOSE 8443

CMD ["./server"]