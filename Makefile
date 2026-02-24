# ==========================================
# Restless – Clean Production Makefile
# ==========================================

APP        := restless
CMD        := ./cmd/restless
DIST       := dist
PKG        := github.com/bspippi1337/restless/internal/version

RAW_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)
VERSION := $(shell echo $(RAW_VERSION) | sed 's/^v//; s/-/+git/; s/-/./g')
PREFIX     ?= /usr/local
BINDIR     := $(PREFIX)/bin

APT_KEY    ?= 5EA836A98EB9E38A51466BF6A0CB94CCA7E69627
APT_DIST   ?= stable
APT_COMP   ?= main
APT_ARCH   ?= amd64

LDFLAGS    := -ldflags "-X $(PKG).Version=$(VERSION)"

.PONY: all build build-all linux darwin windows \
        install uninstall clean test tidy fmt doctor \
        deb publish-apt version teacher

# ------------------------------------------
# Default
# ------------------------------------------

all: build

# ------------------------------------------
# Build
# ------------------------------------------

build:
	@echo "==> Building $(APP) ($(VERSION))"
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
	@echo "==> Cross builds complete"

# ------------------------------------------
# Debian package (proper)
# ------------------------------------------

DEB_REV     ?= 1
ARCH        := $(shell dpkg --print-architecture)
RAW_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.0.0)
BASE_VER    := $(shell echo $(RAW_VERSION) | sed 's/^v//; s/-/+git/; s/-/./g')
DEB_VERSION := $(BASE_VER)-$(DEB_REV)

deb: build
	@echo "==> Building Debian package $(DEB_VERSION) ($(ARCH))"

	rm -rf build/deb
	mkdir -p build/deb/DEBIAN
	mkdir -p build/deb/usr/bin
	mkdir -p build/deb/usr/share/doc/$(APP)

	cp $(APP) build/deb/usr/bin/$(APP)

	echo "Package: $(APP)" > build/deb/DEBIAN/control
	echo "Version: $(DEB_VERSION)" >> build/deb/DEBIAN/control
	echo "Section: utils" >> build/deb/DEBIAN/control
	echo "Priority: optional" >> build/deb/DEBIAN/control
	echo "Architecture: $(ARCH)" >> build/deb/DEBIAN/control
	echo "Maintainer: bspippi1337" >> build/deb/DEBIAN/control
	echo "Depends: libc6 (>= 2.31)" >> build/deb/DEBIAN/control
	echo "Description: Terminal-first API Workbench" >> build/deb/DEBIAN/control

	echo "Restless ($(DEB_VERSION)) unstable; urgency=medium" > build/deb/usr/share/doc/$(APP)/changelog
	echo "" >> build/deb/usr/share/doc/$(APP)/changelog
	echo "  * Automated build" >> build/deb/usr/share/doc/$(APP)/changelog
	echo "" >> build/deb/usr/share/doc/$(APP)/changelog
	echo " -- bspippi1337 <noreply@github.com>  $$(date -R)" >> build/deb/usr/share/doc/$(APP)/changelog

	gzip -9 build/deb/usr/share/doc/$(APP)/changelog

	dpkg-deb --build build/deb $(APP)_$(DEB_VERSION)_$(ARCH).deb
# ------------------------------------------
# Signed APT publish
# ------------------------------------------

publish-apt: deb
	@echo "==> Publishing signed APT repository"

	mkdir -p apt/pool/$(APT_COMP)/r/restless
	mkdir -p apt/dists/$(APT_DIST)/$(APT_COMP)/binary-$(APT_ARCH)

	@DEB=$$(ls -1 $(APP)_*_$(APT_ARCH).deb | sort -V | tail -n1); \
	mv -f $$DEB apt/pool/$(APT_COMP)/r/restless/

	cd apt && \
	dpkg-scanpackages pool /dev/null > dists/$(APT_DIST)/$(APT_COMP)/binary-$(APT_ARCH)/Packages && \
	gzip -9c dists/$(APT_DIST)/$(APT_COMP)/binary-$(APT_ARCH)/Packages > dists/$(APT_DIST)/$(APT_COMP)/binary-$(APT_ARCH)/Packages.gz && \
	apt-ftparchive release dists/$(APT_DIST) > dists/$(APT_DIST)/Release && \
	gpg --batch --yes --default-key $(APT_KEY) --clearsign -o dists/$(APT_DIST)/InRelease dists/$(APT_DIST)/Release && \
	gpg --batch --yes --default-key $(APT_KEY) -abs -o dists/$(APT_DIST)/Release.gpg dists/$(APT_DIST)/Release && \
	gpg --armor --export $(APT_KEY) > restless.gpg

	git add apt
	git commit -m "APT: publish $(VERSION)" || true
	git push

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

teacher: build
	./$(APP) teacher

clean:
	rm -f $(APP)
	rm -rf $(DIST)
	rm -rf build

version:
	@echo $(VERSION)
.PHONY: release-all

release-all: clean build-all deb publish-apt
	@echo "==> Full release pipeline complete."
