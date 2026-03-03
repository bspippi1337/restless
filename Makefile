SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := all

APP ?= $(shell ls -1 cmd 2>/dev/null | head -n1)
ifeq ($(strip $(APP)),)
$(error No cmd/<app> found. Set APP=<name>)
endif

PKG      := ./cmd/$(APP)
OUT      := build
DIST     := dist
PREFIX   ?= /usr/local
BINDIR   := $(PREFIX)/bin

VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS  := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

BIN      := $(OUT)/$(APP)

OSES     := linux darwin windows
ARCHS    := amd64 arm64

.PHONY: all
all: clean build man completions install dist checksums

.PHONY: build
build:
	mkdir -p $(OUT)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) $(PKG)

.PHONY: man
man:
	mkdir -p $(OUT)/man
	if [ -f docs/man/$(APP).1.scd ]; then \
		scdoc < docs/man/$(APP).1.scd > $(OUT)/man/$(APP).1; \
	elif [ -f docs/man/$(APP).1.md ]; then \
		pandoc -s -t man docs/man/$(APP).1.md -o $(OUT)/man/$(APP).1; \
	else \
		$(BIN) --help 2>/dev/null | sed '1i.TH $(APP) 1' > $(OUT)/man/$(APP).1 || true; \
	fi
	gzip -f $(OUT)/man/$(APP).1 || true

.PHONY: completions
completions:
	mkdir -p $(OUT)/completions
	$(BIN) completion bash > $(OUT)/completions/$(APP).bash 2>/dev/null || true
	$(BIN) completion zsh  > $(OUT)/completions/_$(APP) 2>/dev/null || true
	$(BIN) completion fish > $(OUT)/completions/$(APP).fish 2>/dev/null || true

.PHONY: install
install:
	install -d $(BINDIR)
	install -m 0755 $(BIN) $(BINDIR)/$(APP)

.PHONY: dist
dist:
	rm -rf $(DIST)
	mkdir -p $(DIST)
	for os in $(OSES); do \
	  for arch in $(ARCHS); do \
	    ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
	    dir="$(DIST)/$(APP)_$(VERSION)_$${os}_$${arch}"; \
	    mkdir -p "$$dir"; \
	    CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" \
	      go build -trimpath -ldflags "$(LDFLAGS)" -o "$$dir/$(APP)$$ext" $(PKG); \
	    cp -f README* LICENSE "$$dir/" 2>/dev/null || true; \
	    tar -C "$(DIST)" -czf "$$dir.tar.gz" "$$(basename "$$dir")"; \
	    rm -rf "$$dir"; \
	  done; \
	done

.PHONY: checksums
checksums:
	cd $(DIST) && (sha256sum *.tar.gz > SHA256SUMS 2>/dev/null || shasum -a 256 *.tar.gz > SHA256SUMS || true)

.PHONY: clean
clean:
	rm -rf $(OUT) $(DIST)
