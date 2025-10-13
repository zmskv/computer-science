FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shortener cmd/shortener/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/shortener /app/shortener

COPY --from=builder /app/migrations /app/migrations

WORKDIR /app

EXPOSE 8080

CMD ["./shortener"]
