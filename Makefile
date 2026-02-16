SHELL := /bin/sh

APP := restless
DIST := dist

.PHONY: test build clean

test:
	@CGO_ENABLED=0 go test ./... -count=1 -tags "netgo osusergo"

build:
	@mkdir -p $(DIST)
	@CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -tags "netgo osusergo" -o $(DIST)/$(APP) ./cmd/restless
	@echo "Built: $(DIST)/$(APP)"

clean:
	@rm -rf $(DIST)
