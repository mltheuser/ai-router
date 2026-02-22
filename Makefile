.PHONY: build test test-integration test-integration-openrouter test-integration-ollama clean

BINARY_NAME=ai-router
BUILD_DIR=./bin

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ai-router

test:
	go test ./... -v -count=1

test-integration-openrouter:
	go test ./providers/openrouter/ -tags=integration -v -timeout 2m -count=1

test-integration-ollama:
	go test ./providers/ollama/ -tags=integration -v -timeout 2m -count=1

test-integration: test-integration-openrouter test-integration-ollama

clean:
	rm -rf $(BUILD_DIR)
