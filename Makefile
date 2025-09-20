APP := bankapp
PKG := ./...
CMD := ./cmd/bankapp

.PHONY: tidy build test lint run cover

tidy:
	go mod tidy

build:
	go build $(PKG)

test:
	go test $(PKG) -race -cover

lint:
	golangci-lint run

run:
	go run $(CMD)

cover:
	go test $(PKG) -coverprofile=coverage.out
	go tool cover -func=coverage.out
