.PHONY: build generate test

build:
	@go build -o skiff main.go

generate:
	@go generate ./...

test: generate
	@go test ./...

