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
	@echo "Running all tests (excluding performance)..."
	@go test ./... -v

test-unit:
	@echo "Running unit tests..."
	@go test ./internal/handler/... -v

test-component:
	@echo "Running component tests..."
	@go test ./tests/component/... -v

test-dependency:
	@echo "Running dependency tests..."
	@go test ./tests/dependency/... -v

test-performance:
	@echo "Running performance tests..."
	@go test -bench=. ./tests/performance/... -v

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
