BINARY := restless

.PHONY: build test clean

build:
	CGO_ENABLED=0 go build -o $(BINARY) ./cmd/restless

test:
	go test ./...

clean:
	rm -f $(BINARY)
