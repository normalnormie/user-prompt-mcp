.PHONY: build build-client build-server test clean fmt vet lint all install-client install-server

CLIENT_BINARY_NAME=user-prompt-mcp
SERVER_BINARY_NAME=user-prompt-server
BUILD_DIR=bin

all: test build

build: build-client build-server

build-client:
	@echo "Building client ($(CLIENT_BINARY_NAME))..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) ./cmd/user-prompt-mcp

build-server:
	@echo "Building server ($(SERVER_BINARY_NAME))..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(SERVER_BINARY_NAME) ./cmd/user-prompt-server

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

install-client: build-client
	@echo "Installing client ($(CLIENT_BINARY_NAME))..."
	@cp $(BUILD_DIR)/$(CLIENT_BINARY_NAME) $(shell go env GOPATH)/bin/
	@echo "$(CLIENT_BINARY_NAME) installed to $(shell go env GOPATH)/bin/"

install-server: build-server
	@echo "Installing server ($(SERVER_BINARY_NAME))..."
	@cp $(BUILD_DIR)/$(SERVER_BINARY_NAME) $(shell go env GOPATH)/bin/
	@echo "$(SERVER_BINARY_NAME) installed to $(shell go env GOPATH)/bin/"

# A simple way to run the server for testing (optional)
run-server: build-server
	@echo "Running server ($(SERVER_BINARY_NAME))..."
	@$(BUILD_DIR)/$(SERVER_BINARY_NAME)

# Note: Running the client directly is usually not how it's used;
# Cursor runs it.
