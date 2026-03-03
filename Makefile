APP        := restless
CMD        := ./cmd/restless
DIST       := dist
VERSION    := $(shell git describe --tags --always --dirty)
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

.PHONY: help build run clean fmt tidy vet test release release-linux release-darwin release-windows

help:
	@echo "restless build system"
	@echo
	@echo "make build           build binary"
	@echo "make run             run CLI"
	@echo "make clean           remove build artifacts"
	@echo "make fmt             format code"
	@echo "make tidy            tidy go modules"
	@echo "make vet             run go vet"
	@echo "make test            run tests"
	@echo "make release         build cross-platform binaries"
	@echo

build:
	go build $(LDFLAGS) -o $(APP) $(CMD)

run:
	go run $(CMD)

clean:
	rm -rf $(APP) $(DIST)

fmt:
	go fmt ./...

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

release: clean
	mkdir -p $(DIST)
	$(MAKE) release-linux
	$(MAKE) release-darwin
	$(MAKE) release-windows
	@echo "binaries written to $(DIST)/"

release-linux:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP)-linux-amd64 $(CMD)

release-darwin:
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP)-darwin-amd64 $(CMD)

release-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(APP)-windows-amd64.exe $(CMD)
