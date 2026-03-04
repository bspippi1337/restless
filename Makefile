# restless — GNU-style developer Makefile
# ---------------------------------------

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
MANDIR ?= $(PREFIX)/share/man/man1

APP      := restless
SRC      := ./cmd/restless
BUILD    := build
DIST     := dist

GO       := go

VERSION  := $(shell git describe --tags --dirty --always 2>/dev/null || echo dev)
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

.PHONY: all build clean install uninstall man release deb doctor demo

# ---------------------------------------------------------------------

all: build

build:
	mkdir -p $(BUILD)
	CGO_ENABLED=0 $(GO) build \
	-trimpath \
	-ldflags "$(LDFLAGS)" \
	-o $(BUILD)/$(APP) \
	$(SRC)

# ---------------------------------------------------------------------

man:
	mkdir -p $(BUILD)/man
	if [ -f docs/man/restless.1.scd ]; then \
		scdoc < docs/man/restless.1.scd > $(BUILD)/man/restless.1; \
	elif [ -f docs/man/restless.1.md ]; then \
		pandoc -s -t man docs/man/restless.1.md -o $(BUILD)/man/restless.1; \
	else \
		$(BUILD)/$(APP) --help 2>/dev/null | sed '1i.TH restless 1' > $(BUILD)/man/restless.1 || true; \
	fi

# ---------------------------------------------------------------------

install: build man
	install -d $(DESTDIR)$(BINDIR)
	install -m 0755 $(BUILD)/$(APP) $(DESTDIR)$(BINDIR)/$(APP)

	install -d $(DESTDIR)$(MANDIR)
	install -m 0644 $(BUILD)/man/restless.1 $(DESTDIR)$(MANDIR)/restless.1

	@echo "✓ installed $(APP) -> $(DESTDIR)$(BINDIR)/$(APP)"

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(APP)
	rm -f $(DESTDIR)$(MANDIR)/restless.1
	@echo "✓ removed $(APP)"

# ---------------------------------------------------------------------

release: clean
	mkdir -p $(DIST)

	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		out="$(DIST)/$(APP)-$$GOOS-$$GOARCH"; \
		echo "→ $$GOOS/$$GOARCH"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build \
			-trimpath \
			-ldflags "$(LDFLAGS)" \
			-o $$out \
			$(SRC); \
	done

	@echo "✓ release artifacts in $(DIST)/"

# ---------------------------------------------------------------------

deb: build
	mkdir -p $(DIST)/deb/DEBIAN
	mkdir -p $(DIST)/deb/usr/bin

	cp $(BUILD)/$(APP) $(DIST)/deb/usr/bin/

	printf "Package: restless\nVersion: %s\nArchitecture: amd64\nMaintainer: Restless\nDescription: API discovery tool\n" "$(VERSION)" > $(DIST)/deb/DEBIAN/control

	dpkg-deb --build $(DIST)/deb $(DIST)/$(APP)_$(VERSION)_amd64.deb

	@echo "✓ built $(DIST)/$(APP)_$(VERSION)_amd64.deb"

# ---------------------------------------------------------------------

doctor:
	@echo "== restless doctor =="
	@echo

	@printf "go:      "
	@which go || echo "missing"

	@printf "git:     "
	@which git || echo "missing"

	@printf "curl:    "
	@which curl || echo "missing"

	@printf "jq:      "
	@which jq || echo "optional"

	@echo
	@echo "repo:"
	@git status --short || true
	@echo
	@echo "version: $(VERSION)"

# ---------------------------------------------------------------------

demo: build
	@echo
	@echo "== restless demo =="
	@echo

	@echo "scan github api"
	@$(BUILD)/$(APP) scan https://api.github.com || true
	@echo
	@echo "map"
	@$(BUILD)/$(APP) map || true
	@echo
	@echo "inspect"
	@$(BUILD)/$(APP) inspect || true

# ---------------------------------------------------------------------

clean:
	rm -rf $(BUILD) $(DIST)
