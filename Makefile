.PHONY: all build test clean run

BINARY_NAME=mcpop
BUILD_DIR=bin

all: build

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/mcpop

test:
	go test -v ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR)

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)
