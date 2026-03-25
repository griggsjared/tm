.PHONY: test build lint fmt

test:
	go test -v -cover ./...

build:
	go build -o tm .

lint:
	golangci-lint run

fmt:
	go fmt ./...
