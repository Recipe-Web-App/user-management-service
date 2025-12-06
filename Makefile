# Simple Makefile for a Go project

build:
	@echo "Building..."
	@go build -o bin/server cmd/api/main.go

run:
	@go run cmd/api/main.go

clean:
	@echo "Cleaning..."
	@rm -rf bin
	@rm -f main

test:
	@echo "Testing..."
	@go test ./... -v

.PHONY: build run clean test
