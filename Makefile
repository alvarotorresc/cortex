BINARY_NAME := cortex
BUILD_DIR := ./bin
CMD_DIR := ./cmd/cortex

.PHONY: build run test lint fmt clean

## build: Compile the Cortex binary
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

## run: Build and run the Cortex server
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## test: Run all tests with race detection
test:
	go test -race -count=1 ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format all Go source files
fmt:
	gofmt -w .
	goimports -w .

## clean: Remove build artifacts and runtime data
clean:
	rm -rf $(BUILD_DIR)
	go clean -cache

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
