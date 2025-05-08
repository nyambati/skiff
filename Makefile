.PHONY: build generate test

build:
	@go build -o bin/skiff cmd/main.go

generate:
	@go generate ./...

test: generate
	@go test ./...
