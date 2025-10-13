.PHONY: help build run test clean migrate

build:
	go build -o bin/shortener cmd/shortener/main.go

run:
	go run cmd/shortener/main.go

test:
	go test ./...

lint:
	golangci-lint run

