APP := restless
BIN := build
CMD := ./cmd/restless
PREFIX ?= $(HOME)/.local
BINDIR := $(PREFIX)/bin
MANDIR := $(PREFIX)/share/man/man1

GO := go

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X github.com/bspippi1337/restless/internal/cli.buildVersion=$(VERSION) \
           -X github.com/bspippi1337/restless/internal/cli.buildCommit=$(COMMIT) \
           -X github.com/bspippi1337/restless/internal/cli.buildDate=$(DATE)

.DEFAULT_GOAL := help

## help
help:
	@echo ""
	@echo "Restless build system"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //'

## build
build:
	mkdir -p $(BIN)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN)/$(APP) $(CMD)

## run
run:
	$(GO) run $(CMD)

## man
man:
	mkdir -p $(MANDIR)
	printf ".Dd May 2, 2026\n.Dt RESTLESS 1\n.Os\n.Sh NAME\n.Nm restless\n.Nd unix-style file watcher\n" > $(MANDIR)/$(APP).1

## install
install: build man
	mkdir -p $(BINDIR)
	cp $(BIN)/$(APP) $(BINDIR)/$(APP)
	chmod +x $(BINDIR)/$(APP)
	@echo "Installed → $(BINDIR)/$(APP)"
	@echo "Run:"
	@echo "export PATH=\"$(BINDIR):\$$PATH\""

## clean
clean:
	rm -rf $(BIN)

## fmt
fmt:
	$(GO) fmt ./...

## vet
vet:
	$(GO) vet ./...

## test
test:
	$(GO) test ./...

## tidy
tidy:
	$(GO) mod tidy

## doctor
doctor: fmt vet tidy build
	@echo "repo healthy"
