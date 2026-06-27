.PHONY: build test clean

BINARY_NAME=ai-router
BUILD_DIR=./bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ai-router

test:
	go test ./... -v -count=1

clean:
	rm -rf $(BUILD_DIR)
