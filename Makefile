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

lint:
	@echo "Running pre-commit hooks..."
	@pre-commit run --all-files

check: lint test build

.PHONY: build run clean test lint check
