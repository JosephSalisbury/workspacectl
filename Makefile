.PHONY: build test lint fmt check install clean

build:
	go build -o workspacectl .

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

check: build lint test

install:
	go install .

clean:
	rm -f workspacectl
