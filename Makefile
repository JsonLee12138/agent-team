BINARY := agent-team
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build test lint clean

build:
	go build -ldflags "-X github.com/leeforge/agent-team/cmd.Version=$(VERSION)" -o $(BINARY) .

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
