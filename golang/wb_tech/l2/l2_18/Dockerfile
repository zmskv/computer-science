FROM golang:1.24.2 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/app/main.go


FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y curl make && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/order-service /app/order-service

WORKDIR /app


EXPOSE 8080


CMD ["./calendar-service"]