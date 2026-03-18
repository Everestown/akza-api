.PHONY: run build migrate-up migrate-down migrate-create lint test

DB_URL ?= $(shell grep -A1 'database:' config.yaml | tail -1 | tr -d ' url:"')

run:
	go run ./cmd/server/...

build:
	go build -o bin/server ./cmd/server/...

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	goose -dir migrations create $(name) sql

lint:
	golangci-lint run ./...

test:
	go test ./... -v -count=1

tidy:
	go mod tidy
