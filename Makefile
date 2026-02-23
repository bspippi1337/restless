BINARY := restless
PREFIX ?= $(shell echo $/data/data/com.termux/files/usr)
BINDIR := $(PREFIX)/bin

.PHONY: all build install clean

all: build

build:
	CGO_ENABLED=0 go build -o $(BINARY) ./cmd/restless

install: build
	mkdir -p $(BINDIR)
	cp $(BINARY) $(BINDIR)/$(BINARY)
	chmod +x $(BINDIR)/$(BINARY)

clean:
	rm -f $(BINARY)
