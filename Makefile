# Simple Makefile for a Go project

build:
	@echo "Building..."
	@go build -o bin/server cmd/api/main.go

run:
	@if [ -f .env.local ]; then \
		echo "Loading .env.local..."; \
		export $$(grep -v '^#' .env.local | xargs); \
	fi; \
	go run cmd/api/main.go

clean:
	@echo "Cleaning..."
	@rm -rf bin
	@rm -f main

test:
	@echo "Running all tests (excluding performance)..."
	@if [ -f .env.local ]; then \
		echo "Loading .env.local..."; \
		export $$(grep -v '^#' .env.local | xargs); \
	fi; \
	go test ./... -v

test-unit:
	@echo "Running unit tests..."
	@if [ -f .env.local ]; then \
		echo "Loading .env.local..."; \
		export $$(grep -v '^#' .env.local | xargs); \
	fi; \
	go test ./internal/handler/... -v

test-component:
	@echo "Running component tests..."
	@if [ -f .env.local ]; then \
		echo "Loading .env.local..."; \
		export $$(grep -v '^#' .env.local | xargs); \
	fi; \
	go test ./tests/component/... -v

test-dependency:
	@echo "Running dependency tests..."
	@if [ -f .env.local ]; then \
		echo "Loading .env.local..."; \
		export $$(grep -v '^#' .env.local | xargs); \
	fi; \
	go test ./tests/dependency/... -v

test-performance:
	@echo "Running performance tests..."
	@if [ -f .env.local ]; then \
		echo "Loading .env.local..."; \
		export $$(grep -v '^#' .env.local | xargs); \
	fi; \
	go test -bench=. ./tests/performance/... -v

test-all: test-unit test-component test-dependency test-performance

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

view-coverage:
	@echo "Viewing coverage report..."
	@open coverage.html

lint:
	@echo "Running pre-commit hooks..."
	@pre-commit run --all-files

check: lint test build

.PHONY: build run clean test lint check test-unit test-component test-dependency test-performance test-all test-coverage
