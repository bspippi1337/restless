# Restless - all-in distribution Makefile
APP := restless
PKG := github.com/bspippi1337/restless

PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
MANDIR := $(PREFIX)/share/man/man1
BASHDIR := $(PREFIX)/share/bash-completion/completions
ZSHDIR := $(PREFIX)/share/zsh/site-functions

BUILD := build
DIST := dist
BIN := $(BUILD)/$(APP)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
 -X $(PKG)/internal/cli.version=$(VERSION)

GOFLAGS := -trimpath

# Docker
IMAGE ?= restless
DOCKER_TAG ?= $(VERSION)

.PHONY: build clean install uninstall completion man release doctor docker docker-run deb aptrepo

build:
	mkdir -p $(BUILD)
	CGO_ENABLED=0 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/restless

clean:
	rm -rf $(BUILD) $(DIST) .deb-build .apt-repo

completion: build
	mkdir -p $(DIST)/completion
	$(BIN) completion --out $(DIST)/completion

man: build
	mkdir -p $(DIST)/man
	if [ -f docs/man/restless.1.scd ]; then \
		scdoc < docs/man/restless.1.scd > $(DIST)/man/restless.1 ; \
	elif [ -f docs/man/restless.1.md ]; then \
		pandoc -s -t man docs/man/restless.1.md -o $(DIST)/man/restless.1 ; \
	else \
		$(BIN) --help | sed '1i.TH restless 1' > $(DIST)/man/restless.1 ; \
	fi

install: build completion man
	install -Dm755 $(BIN) $(DESTDIR)$(BINDIR)/$(APP)
	install -Dm644 $(DIST)/completion/restless.bash $(DESTDIR)$(BASHDIR)/restless || true
	install -Dm644 $(DIST)/completion/_restless $(DESTDIR)$(ZSHDIR)/_restless || true
	install -Dm644 $(DIST)/man/restless.1 $(DESTDIR)$(MANDIR)/restless.1 || true
	@echo "Installed $(APP) -> $(DESTDIR)$(BINDIR)/$(APP)"

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(APP)
	rm -f $(DESTDIR)$(BASHDIR)/restless
	rm -f $(DESTDIR)$(ZSHDIR)/_restless
	rm -f $(DESTDIR)$(MANDIR)/restless.1

release: clean
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64   go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)-linux-amd64 ./cmd/restless
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64   go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)-linux-arm64 ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)-darwin-amd64 ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64  go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)-darwin-arm64 ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(DIST)/$(APP)-windows-amd64.exe ./cmd/restless

doctor:
	@echo "Go: $$(go version)"
	@echo "Git version: $(VERSION)"
	go build ./...
	go test ./... || true

docker:
	docker build -t $(IMAGE):$(DOCKER_TAG) .

docker-run: docker
	docker run --rm $(IMAGE):$(DOCKER_TAG) --help

deb: build
	./scripts/build_deb.sh $(VERSION)

aptrepo: deb
	./scripts/build_apt_repo.sh $(VERSION)
