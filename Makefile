APP := restless
BIN := build
CMD := ./cmd/restless

PREFIX ?= /usr/local
DESTDIR ?=
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man/man1
DOCDIR ?= $(PREFIX)/share/doc/$(APP)

GO ?= go
INSTALL ?= install
PKGS = $(shell $(GO) list ./... 2>/dev/null | grep -v '/archive/' | grep -v '/cmd/wasm')

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X github.com/bspippi1337/restless/internal/cli.buildVersion=$(VERSION) \
           -X github.com/bspippi1337/restless/internal/cli.buildCommit=$(COMMIT) \
           -X github.com/bspippi1337/restless/internal/cli.buildDate=$(DATE)

.DEFAULT_GOAL := help

.PHONY: help build run man install uninstall clean fmt vet test tidy modcheck check doctor list-packages

## help
help:
	@echo ""
	@echo "Restless build system"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //'

## list-packages
list-packages:
	@printf '%s\n' $(PKGS)

## build
build:
	mkdir -p $(BIN)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN)/$(APP) $(CMD)

## run
run:
	$(GO) run $(CMD)

## man
man:
	mkdir -p $(BIN)/man
	printf ".Dd May 8, 2026\n.Dt RESTLESS 1\n.Os\n.Sh NAME\n.Nm restless\n.Nd reactive API discovery and Unix observability runtime\n.Sh SYNOPSIS\n.Nm\n.Op command\n.Op args\n.Sh DESCRIPTION\n.Nm\nperforms safe API surface discovery and reactive filesystem command execution.\n.Sh COMMANDS\n.Bl -tag -width watch\n.It Cm scan\nScan an API target.\n.It Cm learn\nDiscover API endpoints.\n.It Cm map\nRender known endpoint topology.\n.It Cm watch\nWatch a filesystem path and run a shell command on change.\n.El\n.Sh EXAMPLES\n.Nm watch . --run \"make test\"\n" > $(BIN)/man/$(APP).1

## install
install: build man
	$(INSTALL) -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 $(BIN)/$(APP) $(DESTDIR)$(BINDIR)/$(APP)
	$(INSTALL) -d $(DESTDIR)$(MANDIR)
	$(INSTALL) -m 0644 $(BIN)/man/$(APP).1 $(DESTDIR)$(MANDIR)/$(APP).1
	$(INSTALL) -d $(DESTDIR)$(DOCDIR)
	$(INSTALL) -m 0644 README.md $(DESTDIR)$(DOCDIR)/README.md
	@if [ -f COPYING ]; then $(INSTALL) -m 0644 COPYING $(DESTDIR)$(DOCDIR)/COPYING; fi
	@echo "Installed -> $(DESTDIR)$(BINDIR)/$(APP)"

## uninstall
uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(APP)
	rm -f $(DESTDIR)$(MANDIR)/$(APP).1
	rm -rf $(DESTDIR)$(DOCDIR)

## clean
clean:
	rm -rf $(BIN)

## fmt
fmt:
	@test -n "$(PKGS)" || { echo "no supported Go packages found"; exit 1; }
	$(GO) fmt $(PKGS)

## vet
vet:
	@test -n "$(PKGS)" || { echo "no supported Go packages found"; exit 1; }
	$(GO) vet $(PKGS)

## test
test:
	@test -n "$(PKGS)" || { echo "no supported Go packages found"; exit 1; }
	$(GO) test $(PKGS)

## tidy
tidy:
	$(GO) mod tidy

## modcheck
modcheck:
	$(GO) mod tidy
	git diff --exit-code -- go.mod go.sum

## check
check: modcheck fmt vet test build
	@echo "repo healthy"

## doctor
doctor: check
