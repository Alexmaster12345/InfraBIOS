BIN_SERVER := bin/infrabios-server
BIN_AGENT  := bin/infrabios-agent
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -ldflags "-X main.version=$(VERSION) -s -w"

.PHONY: all build server agent run migrate tidy lint test clean help

all: build

## build: compile server and agent binaries
build: server agent

## server: compile the API server
server:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BIN_SERVER) ./cmd/server

## agent: compile the per-server agent
agent:
	@mkdir -p bin
	go build $(LDFLAGS) -o $(BIN_AGENT) ./cmd/agent

## run: start the server (requires PostgreSQL)
run: server
	./$(BIN_SERVER)

## migrate: apply database migrations
migrate:
	psql "$$INFRABIOS_DATABASE_URL" -f internal/migrations/001_init.sql

## tidy: tidy Go modules
tidy:
	go mod tidy

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## test: run all tests
test:
	go test -race -count=1 ./...

## clean: remove build artefacts
clean:
	rm -rf bin/

## help: list targets
help:
	@grep -E '^##' Makefile | sed 's/## //'
