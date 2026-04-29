# Variables
-include .env

GOPATH=$(shell go env GOPATH)
SWAG_BIN=$(GOPATH)/bin/swag
BINARY=bin/mirandaclin
MAIN=./cmd/api

.PHONY: build test lint swagger clean up down

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY) $(MAIN)

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

up:
	make swagger
	docker compose up --build -d

down:
	@echo "Encerrando containers..."
	docker compose down

clean:
	@echo "Cleaning up..."
	rm -rf bin/ docs/
