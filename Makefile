# ==========================================
# Restless – Production Makefile
# ==========================================

APP        := restless
CMD        := ./cmd/restless
DIST       := dist
PKG        := github.com/bspippi1337/restless/internal/version

PREFIX     ?= /usr/local
BINDIR     := $(PREFIX)/bin

DEB_REV    ?= 1
ARCH       := $(shell dpkg --print-architecture)

# ------------------------------------------
# Version logic (Debian correct)
# ------------------------------------------

GIT_TAG    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
COMMITS_SINCE_TAG := $(shell if [ -n "$(GIT_TAG)" ]; then git rev-list $(GIT_TAG)..HEAD --count; else git rev-list HEAD --count; fi)

BASE_VER   := $(shell echo $(GIT_TAG) | sed 's/^v//')
ifeq ($(BASE_VER),)
BASE_VER := 0.0.0
endif

ifeq ($(COMMITS_SINCE_TAG),0)
UPSTREAM_VER := $(BASE_VER)
else
UPSTREAM_VER := $(BASE_VER)+git$(COMMITS_SINCE_TAG)
endif

DEB_VERSION := $(UPSTREAM_VER)-$(DEB_REV)

LDFLAGS := -ldflags "-X $(PKG).Version=$(UPSTREAM_VER)"

.PHONY: all build build-all linux darwin windows \
        install uninstall clean test tidy fmt doctor \
        deb publish-apt changelog version

# ------------------------------------------
# Default
# ------------------------------------------

all: build

# ------------------------------------------
# Build
# ------------------------------------------

build:
	@echo "==> Building $(APP) $(UPSTREAM_VER)"
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(APP) $(CMD)

linux:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP)_linux_amd64 $(CMD)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(APP)_linux_arm64 $(CMD)

darwin:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP)_darwin_amd64 $(CMD)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(APP)_darwin_arm64 $(CMD)

windows:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP)_windows_amd64.exe $(CMD)

build-all: linux darwin windows

# ------------------------------------------
# Changelog (auto from git)
# ------------------------------------------

changelog:
	@echo "==> Generating changelog from git"
	rm -f build/changelog
	mkdir -p build
	echo "$(APP) ($(DEB_VERSION)) unstable; urgency=medium" > build/changelog
	echo "" >> build/changelog
	git log --pretty=format:"  * %s" $(GIT_TAG)..HEAD >> build/changelog || true
	echo "" >> build/changelog
	echo " -- bspippi1337 <noreply@github.com>  $$(date -R)" >> build/changelog
	gzip -9 build/changelog

# ------------------------------------------
# Debian package
# ------------------------------------------

deb: build changelog
	@echo "==> Building Debian package $(DEB_VERSION)"

	rm -rf build/deb
	mkdir -p build/deb/DEBIAN
	mkdir -p build/deb/usr/bin
	mkdir -p build/deb/usr/share/doc/$(APP)

	cp $(APP) build/deb/usr/bin/$(APP)
	cp build/changelog.gz build/deb/usr/share/doc/$(APP)/changelog.gz

	echo "Package: $(APP)" > build/deb/DEBIAN/control
	echo "Version: $(DEB_VERSION)" >> build/deb/DEBIAN/control
	echo "Section: utils" >> build/deb/DEBIAN/control
	echo "Priority: optional" >> build/deb/DEBIAN/control
	echo "Architecture: $(ARCH)" >> build/deb/DEBIAN/control
	echo "Maintainer: bspippi1337" >> build/deb/DEBIAN/control
	echo "Depends: libc6 (>= 2.31)" >> build/deb/DEBIAN/control
	echo "Description: Terminal-first API Workbench" >> build/deb/DEBIAN/control

	dpkg-deb --build build/deb $(APP)_$(DEB_VERSION)_$(ARCH).deb

# ------------------------------------------
# Install
# ------------------------------------------

install: build
	install -d $(BINDIR)
	install -m 0755 $(APP) $(BINDIR)/$(APP)

uninstall:
	rm -f $(BINDIR)/$(APP)

# ------------------------------------------
# Dev tools
# ------------------------------------------

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

doctor: tidy fmt test build
	@echo "Doctor OK"

clean:
	rm -rf build $(DIST) $(APP)

version:
	@echo $(DEB_VERSION)
