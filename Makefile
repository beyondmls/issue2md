.PHONY: build test test-unit test-integration clean fmt vet

BIN_DIR := bin
BIN_NAME := issue2md

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) ./cmd/issue2md

test: test-unit test-integration

test-unit:
	go test -v -short ./internal/...

test-integration:
	go test -v ./internal/github/...

clean:
	rm -rf $(BIN_DIR)/

fmt:
	go fmt ./...

vet:
	go vet ./...
