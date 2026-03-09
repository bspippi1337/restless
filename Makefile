kAPP := restless
BIN := build
CMD := ./cmd/restless

GO := go

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X main.version=$(VERSION) \
           -X main.commit=$(COMMIT) \
           -X main.date=$(DATE)

.DEFAULT_GOAL := help

## help: show available targets
help:
	@echo ""
	@echo "Restless build system"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //'

## build: compile binary
build:
	mkdir -p $(BIN)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN)/$(APP) $(CMD)

## run: run without compiling
run:
	$(GO) run $(CMD)

## install: install into go bin
install:
	$(GO) install -ldflags "$(LDFLAGS)" $(CMD)

## clean: remove build artifacts
clean:
	rm -rf $(BIN)

## fmt: format code
fmt:
	$(GO) fmt ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## test: run tests
test:
	$(GO) test ./...

## tidy: fix go.mod
tidy:
	$(GO) mod tidy

## doctor: full repo health check
doctor: fmt vet tidy build
	@echo "repo healthy"

## release: build multi-platform binaries
release:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build -o dist/$(APP)-linux-amd64 $(CMD)
	GOOS=darwin GOARCH=amd64 $(GO) build -o dist/$(APP)-mac-amd64 $(CMD)
	GOOS=windows GOARCH=amd64 $(GO) build -o dist/$(APP)-windows-amd64.exe $(CMD)

## docker: run autorest scanner container
docker:
	docker build -t restless .
