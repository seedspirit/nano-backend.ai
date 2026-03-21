.PHONY: build test lint fmt clean

build:
	go build ./...

test:
	go test ./... -v

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	go clean ./...
