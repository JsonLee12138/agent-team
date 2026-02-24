BINARY := agent-team
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GO := $(shell command -v go1.24.2 2>/dev/null || GOROOT=$$HOME/.gvm/gos/go1.24.2 echo $$HOME/.gvm/gos/go1.24.2/bin/go 2>/dev/null || echo go)
GORUN := GOROOT=$(HOME)/.gvm/gos/go1.24.2 $(HOME)/.gvm/gos/go1.24.2/bin/go

.PHONY: build test lint clean

build:
	$(GORUN) build -ldflags "-X github.com/JsonLee12138/agent-team/cmd.Version=$(VERSION)" -o $(BINARY) .

test:
	$(GORUN) test ./... -v

lint:
	$(GORUN) vet ./...

clean:
	rm -f $(BINARY)
