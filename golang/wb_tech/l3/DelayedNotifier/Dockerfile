FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o notifier-service ./cmd/delayed-notifier/main.go


FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y curl make && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/notifier-service /app/notifier-service

WORKDIR /app


EXPOSE 8080


CMD ["./notifier-service"]