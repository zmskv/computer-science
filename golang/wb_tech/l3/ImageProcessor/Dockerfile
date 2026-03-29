FROM golang:1.25 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /out/image-processor ./cmd/image-processor

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /out/image-processor /app/image-processor
COPY web /app/web
RUN mkdir -p /app/storage/originals /app/storage/processed /app/storage/thumbnails /app/storage/metadata

EXPOSE 8080
CMD ["/app/image-processor"]
