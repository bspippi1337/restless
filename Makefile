BINARY := restless
PKG := ./cmd/restless
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin
DIST := dist

.PHONY: all build run test install uninstall clean release

all: build

build:
	CGO_ENABLED=0 go build -o $(BINARY) $(PKG)

run:
	CGO_ENABLED=0 go run $(PKG)

test:
	go test ./...

install: build
	mkdir -p $(BINDIR)
	install -m 0755 $(BINARY) $(BINDIR)/$(BINARY)

uninstall:
	rm -f $(BINDIR)/$(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

release:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(BINARY)_linux_amd64 $(PKG)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(DIST)/$(BINARY)_darwin_amd64 $(PKG)
