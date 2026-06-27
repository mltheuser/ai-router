.PHONY: build test lint fmt clean

BINARY_NAME=ai-router
BUILD_DIR=./bin
GOLANGCI_LINT=$(BUILD_DIR)/golangci-lint
GOLANGCI_LINT_VERSION=v2.12.2

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ai-router

test:
	go test ./... -v -count=1

# Bootstrap the pinned golangci-lint into ./bin (no-op if already present).
$(GOLANGCI_LINT):
	GOBIN=$(PWD)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

fmt: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) fmt

clean:
	rm -rf $(BUILD_DIR)
