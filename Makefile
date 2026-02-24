➜  restless git:(main) ✗ cat Makefile 
BINARY := restless
VERSION ?= 4.0.4
PKG := github.com/bspippi1337/restless/internal/version

.PHONY: build test clean build-all tidy fmt doctor

build:
	CGO_ENABLED=0 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o $(BINARY) ./cmd/restless

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_linux_amd64   ./cmd/restless
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_linux_arm64   ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_darwin_amd64  ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_darwin_arm64  ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X $(PKG).Version=$(VERSION)" -o dist/restless_windows_amd64.exe ./cmd/restless

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)
	rm -rf dist

doctor: tidy fmt test build
	@echo "Doctor OK"

teacher:
	./restless teacher
# ---- Config ----
APP_NAME := restless
PKG := github.com/bspippi1337/restless
CMD := ./cmd/restless

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin

LDFLAGS := -ldflags "-X $(PKG)/internal/version.Version=$(VERSION)"

# ---- Targets ----

.PHONY: all build install uninstall clean version

all: build

build:
	@echo "==> Building $(APP_NAME) ($(VERSION))"
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(APP_NAME) $(CMD)

install: build
	@echo "==> Installing to $(BINDIR)"
	install -d $(BINDIR)
	install -m 0755 $(APP_NAME) $(BINDIR)/$(APP_NAME)
	@echo "Installed: $(BINDIR)/$(APP_NAME)"

uninstall:
	@echo "==> Removing $(BINDIR)/$(APP_NAME)"
	rm -f $(BINDIR)/$(APP_NAME)

clean:
	@echo "==> Cleaning build artifacts"
	rm -f $(APP_NAME)

version:
	@echo $(VERSION)
