FROM golang:1.26 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /out/sales-tracker ./cmd/sales-tracker

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /out/sales-tracker /app/sales-tracker
COPY web /app/web

EXPOSE 8080
CMD ["/app/sales-tracker"]
