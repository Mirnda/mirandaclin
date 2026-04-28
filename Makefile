BINARY=bin/mirandaclin
MAIN=./cmd/api

.PHONY: build test lint swagger clean

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
	swag init -g $(MAIN)/main.go -o docs

clean:
	@echo "Cleaning up..."
	rm -rf bin/ docs/
