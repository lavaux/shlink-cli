## ─────────────────────────────────────────────
##  shlink-cli  –  Makefile
## ─────────────────────────────────────────────

BINARY      := shlink
MODULE      := github.com/lavaux/shlink-cli
CMD_PKG     := $(MODULE)/cmd/shlink
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
  -X '$(MODULE)/internal/config.Version=$(VERSION)' \
  -X '$(MODULE)/internal/config.Commit=$(COMMIT)' \
  -X '$(MODULE)/internal/config.BuildDate=$(BUILD_DATE)'

# Platforms for cross-compilation
PLATFORMS := \
  linux/amd64 \
  linux/arm64 \
  darwin/amd64 \
  darwin/arm64 \
  windows/amd64

# ─── Primary targets ──────────────────────────

.DEFAULT_GOAL := build

## build: Compile the binary for the current OS/arch
.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(CMD_PKG)

## run: Build and run with --help
.PHONY: run
run: build
	./bin/$(BINARY) --help

## install: Install the binary to GOPATH/bin
.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" $(CMD_PKG)

# ─── Testing & quality ────────────────────────

## test: Run all unit tests
.PHONY: test
test:
	go test -race -coverprofile=coverage.out ./...

## test-verbose: Run tests with verbose output
.PHONY: test-verbose
test-verbose:
	go test -v -race ./...

## coverage: Open HTML coverage report
.PHONY: coverage
coverage: test
	go tool cover -html=coverage.out

## lint: Run golangci-lint (requires golangci-lint in PATH)
.PHONY: lint
lint:
	golangci-lint run ./...

## vet: Run go vet
.PHONY: vet
vet:
	go vet ./...

## fmt: Format all Go source files
.PHONY: fmt
fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true

## check: vet + lint
.PHONY: check
check: vet lint

# ─── Dependency management ────────────────────

## tidy: Tidy go.mod and go.sum
.PHONY: tidy
tidy:
	go mod tidy

## vendor: Vendor dependencies
.PHONY: vendor
vendor:
	go mod vendor

# ─── Cross-compilation / release ──────────────

## dist: Build binaries for all target platforms
.PHONY: dist
dist:
	@mkdir -p dist
	@$(foreach PLATFORM,$(PLATFORMS), \
	  $(eval OS   := $(word 1,$(subst /, ,$(PLATFORM)))) \
	  $(eval ARCH := $(word 2,$(subst /, ,$(PLATFORM)))) \
	  $(eval EXT  := $(if $(filter windows,$(OS)),.exe,)) \
	  echo "→ Building $(OS)/$(ARCH)…"; \
	  GOOS=$(OS) GOARCH=$(ARCH) go build \
	    -ldflags "$(LDFLAGS)" \
	    -o dist/$(BINARY)_$(OS)_$(ARCH)$(EXT) \
	    $(CMD_PKG); \
	)

## dist-checksums: Generate SHA256 checksums for dist artefacts
.PHONY: dist-checksums
dist-checksums: dist
	cd dist && sha256sum * > checksums.txt

# ─── Docker ───────────────────────────────────

## docker-build: Build the Docker image
.PHONY: docker-build
docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  -t $(BINARY):$(VERSION) \
	  -t $(BINARY):latest .

# ─── Housekeeping ─────────────────────────────

## clean: Remove build artefacts
.PHONY: clean
clean:
	rm -rf bin/ dist/ coverage.out

## help: Print this help message
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ": "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' | \
	  sed 's/## //'
