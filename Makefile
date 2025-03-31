.PHONY: build test clean fmt vet lint all install

BINARY_NAME=user-input-mcp
BUILD_DIR=bin

all: test build

build:
	@echo "Building..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/user-input-mcp

test:
	@echo "Testing..."
	@go test -v ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

fmt:
	@echo "Formatting..."
	@go fmt ./...

vet:
	@echo "Vetting..."
	@go vet ./...

lint:
	@echo "Linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed"; \
	fi

install:
	go install ./cmd/user-input-mcp

run: build
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/user-input-mcp