VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
BUILD_LDFLAGS := -X main.version=$(VERSION)

.PHONY: test build lint fmt

test:
	go test -v -cover ./...

build:
	go build -ldflags "$(BUILD_LDFLAGS)" -o tm .

lint:
	go vet ./...
	golangci-lint run

fmt:
	gofmt -s -w .
