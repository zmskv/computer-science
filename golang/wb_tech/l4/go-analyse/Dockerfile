FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o go-analyse ./cmd/go-analyse


FROM debian:bullseye-slim

COPY --from=builder /app/go-analyse /app/go-analyse

WORKDIR /app

EXPOSE 8080

CMD ["./go-analyse"]
