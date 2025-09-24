ifneq (,$(wildcard .env))
    include .env
    export
endif

RUN_DIR = ./cmd/calendar

.PHONY: run 
run:
	go run $(RUN_DIR)/main.go

.PHONY: install
install:
	go mod download

.PHONY : clear
clear:
	go mod tidy

.PHONY: test
test:
	go test -v -count=1 ./...

.PHONY: test100
test100:
	go test -v -count=100 ./...

.PHONY: race
race:
	go test -v -race -count=1 ./...
