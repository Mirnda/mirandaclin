# Variables
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
	@echo "Generating Swagger docs..."
	@echo "If $(SWAG_BIN) is not installed, run: go install github.com/swaggo/swag/cmd/swag@latest"
	$(SWAG_BIN) init -g $(MAIN)/main.go -o docs

up:
	@echo "Subindo infraestrutura (Postgres + Redis)..."
	docker compose up -d postgres redis
	@echo "Aguardando banco ficar pronto..."
	@until docker compose exec postgres pg_isready -U $${DB_USER:-postgres} > /dev/null 2>&1; do sleep 1; done
	@echo "Iniciando aplicação..."
	go run $(MAIN)

down:
	@echo "Encerrando containers..."
	docker compose down

clean:
	@echo "Cleaning up..."
	rm -rf bin/ docs/
