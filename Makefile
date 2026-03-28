.PHONY: build test check clean

build:
	go build ./...

test:
	go test ./... -v

check:
	gofmt -w .
	go build ./...
	golangci-lint run ./...

clean:
	go clean ./...
