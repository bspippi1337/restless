APP := restless
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

help:
	@echo "Restless build system"

build:
	mkdir -p $(BIN)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN)/$(APP) $(CMD)

install:
	$(GO) install -ldflags "$(LDFLAGS)" $(CMD)

clean:
	rm -rf $(BIN)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

man:
	mkdir -p dist/man
	cp docs/man/restless.1 dist/man/

completion:
	mkdir -p dist/completion
	$(BIN)/$(APP) completion bash > dist/completion/restless.bash
	$(BIN)/$(APP) completion zsh > dist/completion/_restless

release:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build -o dist/$(APP)-linux-amd64 $(CMD)
	GOOS=darwin GOARCH=amd64 $(GO) build -o dist/$(APP)-mac-amd64 $(CMD)
	GOOS=windows GOARCH=amd64 $(GO) build -o dist/$(APP)-windows-amd64.exe $(CMD)
