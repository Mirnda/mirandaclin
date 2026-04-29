# Variables
-include .env

GOPATH=$(shell go env GOPATH)
SWAG_BIN=$(GOPATH)/bin/swag
BINARY=bin/mirandaclin
MAIN=./cmd/api

.PHONY: up down clean test lint swagger build lint_doc test_doc

up:
	@echo "Iniciando containers..."
	docker compose up --build -d

down:
	@echo "Encerrando containers..."
	docker compose down

clean:
	@echo "Cleaning up..."
	rm -rf bin/ docs/

test:
	@echo "Running tests..."
	go test -v ./...

lint:
	@echo "Running linter..."
	golangci-lint run

swagger:
ifneq ("${APP_ENV}", "production")
	@echo "Generating Swagger docs..."
	@echo "If $(SWAG_BIN) is not installed, run: go install github.com/swaggo/swag@v1.16.6"
	$(SWAG_BIN) init -g $(MAIN)/main.go -o docs
else
	@echo "Swagger skipped: production environment..."
endif

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY) $(MAIN)

# DOCKER MAKES - For those who do not have Go or GolangCI installed.
GO_IMAGE=golang:1.26-alpine

lint_doc:
	docker run --rm \
		-v $(CURDIR):/app \
		-w /app \
		golangci/golangci-lint:latest \
		golangci-lint run

test_doc:
	docker run --rm \
		-v $(CURDIR):/app \
		-w /app \
		$(GO_IMAGE) \
		go test -v ./...
