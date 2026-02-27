BINARY := output/agent-team
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GO := $(shell command -v go1.24.2 2>/dev/null || GOROOT=$$HOME/.gvm/gos/go1.24.2 echo $$HOME/.gvm/gos/go1.24.2/bin/go 2>/dev/null || echo go)
GORUN := GOROOT=$(HOME)/.gvm/gos/go1.24.2 $(HOME)/.gvm/gos/go1.24.2/bin/go

.PHONY: build test lint clean install uninstall

build:
	$(GORUN) build -ldflags "-X github.com/JsonLee12138/agent-team/cmd.Version=$(VERSION)" -o $(BINARY) .

test:
	$(GORUN) test ./... -v

lint:
	$(GORUN) vet ./...

clean:
	rm -f $(BINARY)

# Install locally: build and symlink to /usr/local/bin for testing
install: build
	@if [ -f /usr/local/bin/agent-team ] && [ ! -L /usr/local/bin/agent-team ]; then \
		echo "⚠  Found non-symlink agent-team at /usr/local/bin/agent-team (likely brew install)"; \
		echo "   Backing up to /usr/local/bin/agent-team.bak"; \
		mv /usr/local/bin/agent-team /usr/local/bin/agent-team.bak; \
	fi
	@mkdir -p /usr/local/bin
	@ln -sf $(CURDIR)/$(BINARY) /usr/local/bin/agent-team
	@echo "✓ Installed agent-team $(VERSION) → /usr/local/bin/agent-team"
	@echo "  (symlinked to $(CURDIR)/$(BINARY))"

# Remove the local symlink and restore backup if exists
uninstall:
	@rm -f /usr/local/bin/agent-team
	@if [ -f /usr/local/bin/agent-team.bak ]; then \
		mv /usr/local/bin/agent-team.bak /usr/local/bin/agent-team; \
		echo "✓ Restored original agent-team from backup"; \
	else \
		echo "✓ Uninstalled agent-team from /usr/local/bin"; \
	fi
