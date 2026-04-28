BINARY=bin/mirandaclin
MAIN=./cmd/api

.PHONY: build test lint swagger clean

build:
	go build -o $(BINARY) $(MAIN)

test:
	go test -v ./...

lint:
	golangci-lint run

swagger:
	swag init -g $(MAIN)/main.go -o docs

clean:
	rm -rf bin/ docs/
