BINARY      = fileoptimizer-agent
AGENT       = ./agent/
LDFLAGS     = -ldflags="-s -w"
WIN_LDFLAGS = -ldflags="-H windowsgui -s -w"
DIST        = ./dist/agents

# Go version used inside Docker (should match go.mod)
GO_IMAGE    = golang:1.25-bookworm

# Flags safe inside Docker bash -c "..." strings.
# Single-quoted value avoids the outer double-quote being closed early.
DOCKER_LDFLAGS = -ldflags='-s -w'

# VCS stamping fails inside Docker (no git history); disable it explicitly.
DOCKER_BUILD_FLAGS = -buildvcs=false

# appindicator was renamed in Debian Bookworm.
APPINDICATOR = libayatana-appindicator3-dev

# Detect host OS/arch at make-time.
GOHOSTOS   := $(shell go env GOHOSTOS)
GOHOSTARCH := $(shell go env GOHOSTARCH)

# Persist the Go module cache between Docker runs to avoid re-downloading
# dependencies on every build.
GOMODCACHE := $(shell go env GOMODCACHE)

.PHONY: all darwin-arm64 darwin-amd64 darwin-universal \
        linux-amd64 linux-arm64 windows deploy clean local \
        server run-server

# ── Default target ────────────────────────────────────────────────────────────
# On macOS build every platform; on Linux skip darwin (requires macOS SDK).

ifeq ($(GOHOSTOS),darwin)
all: darwin-universal linux-amd64 linux-arm64 windows
else
all: linux-amd64 linux-arm64 windows
	@echo "Note: darwin binaries can only be built on macOS."
endif

$(DIST):
	mkdir -p $(DIST)

# ── macOS ─────────────────────────────────────────────────────────────────────
# systray uses Cocoa → CGO_ENABLED=1 required; macOS SDK must be present.
# These targets are silently skipped when the host is not macOS.

darwin-arm64: $(DIST)
ifeq ($(GOHOSTOS),darwin)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) \
		-o $(DIST)/$(BINARY)-darwin-arm64 $(AGENT)
else
	@echo "Skipping darwin-arm64 (host is $(GOHOSTOS), not macOS)"
endif

darwin-amd64: $(DIST)
ifeq ($(GOHOSTOS),darwin)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) \
		-o $(DIST)/$(BINARY)-darwin-amd64 $(AGENT)
else
	@echo "Skipping darwin-amd64 (host is $(GOHOSTOS), not macOS)"
endif

darwin-universal: darwin-arm64 darwin-amd64
ifeq ($(GOHOSTOS),darwin)
	lipo -create -output $(DIST)/$(BINARY)-darwin \
		$(DIST)/$(BINARY)-darwin-arm64 \
		$(DIST)/$(BINARY)-darwin-amd64
	rm $(DIST)/$(BINARY)-darwin-arm64 $(DIST)/$(BINARY)-darwin-amd64
else
	@echo "Skipping darwin-universal (host is $(GOHOSTOS), not macOS)"
endif

# ── Linux amd64 ───────────────────────────────────────────────────────────────
# Native build on Linux/amd64 (fast — no Docker overhead).
# Falls back to Docker when building from macOS or a non-amd64 Linux host.

linux-amd64: $(DIST)
ifeq ($(GOHOSTOS)-$(GOHOSTARCH),linux-amd64)
	@echo "Building linux-amd64 natively..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) \
		-o $(DIST)/$(BINARY)-linux-amd64 $(AGENT)
else
	@echo "Building linux-amd64 via Docker (host is $(GOHOSTOS)/$(GOHOSTARCH))..."
	docker run --rm \
		--platform linux/amd64 \
		-v "$(CURDIR)":/workspace \
		-v "$(GOMODCACHE)":/root/go/pkg/mod \
		-e GOPATH=/root/go \
		-e GOMODCACHE=/root/go/pkg/mod \
		-w /workspace \
		$(GO_IMAGE) \
		bash -c "apt-get update -qq && \
		         apt-get install -y -qq libgtk-3-dev $(APPINDICATOR) 2>/dev/null && \
		         CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
		         go build $(DOCKER_LDFLAGS) $(DOCKER_BUILD_FLAGS) -o $(DIST)/$(BINARY)-linux-amd64 $(AGENT)"
endif

# ── Linux arm64 ───────────────────────────────────────────────────────────────
# Native build on Linux/arm64; Docker with QEMU emulation on all other hosts.

linux-arm64: $(DIST)
ifeq ($(GOHOSTOS)-$(GOHOSTARCH),linux-arm64)
	@echo "Building linux-arm64 natively..."
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) \
		-o $(DIST)/$(BINARY)-linux-arm64 $(AGENT)
else
	@echo "Building linux-arm64 via Docker (host is $(GOHOSTOS)/$(GOHOSTARCH))..."
	docker run --rm \
		--platform linux/arm64 \
		-v "$(CURDIR)":/workspace \
		-v "$(GOMODCACHE)":/root/go/pkg/mod \
		-e GOPATH=/root/go \
		-e GOMODCACHE=/root/go/pkg/mod \
		-w /workspace \
		$(GO_IMAGE) \
		bash -c "apt-get update -qq && \
		         apt-get install -y -qq libgtk-3-dev $(APPINDICATOR) 2>/dev/null && \
		         CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
		         go build $(DOCKER_LDFLAGS) $(DOCKER_BUILD_FLAGS) -o $(DIST)/$(BINARY)-linux-arm64 $(AGENT)"
endif

# ── Windows ───────────────────────────────────────────────────────────────────
# systray on Windows is pure Go → CGO_ENABLED=0 cross-compile works anywhere.

windows: $(DIST)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(WIN_LDFLAGS) \
		-o $(DIST)/$(BINARY)-windows-amd64.exe $(AGENT)

# ── Shortcuts ─────────────────────────────────────────────────────────────────

# macOS + Windows only (no Docker needed)
local: darwin-universal windows

# Build all then restart compose
deploy: all
	docker compose up -d --build

clean:
	rm -rf dist/

# ── Server ────────────────────────────────────────────────────────────────────

server:
	cd server && go build -o server .

run-server:
	cd server && go run .
