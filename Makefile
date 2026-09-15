.PHONY: all build test lint fmt clean

all: test build

build:
	go build -o bin/go-queue .

test:
	go test ./...

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin coverage.out
