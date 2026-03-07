# Hamlet — observability and intelligence for test suites

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test lint clean demo benchmark-fetch benchmark-smoke benchmark-full benchmark-stress benchmark-summary install

# Build the CLI binary
build:
	go build -ldflags "$(LDFLAGS)" -o hamlet ./cmd/hamlet

# Install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/hamlet

# Run all Go tests
test:
	go test ./internal/... ./cmd/...

# Run tests with verbose output
test-v:
	go test -v ./internal/... ./cmd/...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./internal/... ./cmd/...
	go tool cover -func=coverage.out

# Build check (compile only, no binary output)
check:
	go build ./internal/... ./cmd/...

# Clean build artifacts
clean:
	rm -f hamlet coverage.out

# Run demo: analyze the current repository
demo:
	@echo "=== Hamlet Demo ==="
	@echo ""
	@echo "--- hamlet analyze ---"
	go run ./cmd/hamlet analyze
	@echo ""
	@echo "--- hamlet summary ---"
	go run ./cmd/hamlet summary
	@echo ""
	@echo "--- hamlet metrics ---"
	go run ./cmd/hamlet metrics

# Legacy JavaScript tests (requires Node.js 22+)
test-legacy:
	npm test

# ── Public Benchmark Matrix ──────────────────────────────────

# Download benchmark repos (shallow clone by default)
benchmark-fetch:
	./scripts/benchmarks/fetch_public_repos.sh

# Quick benchmark (smoke-tier repos only)
benchmark-smoke:
	./scripts/benchmarks/run_public_matrix.sh smoke
	python3 ./scripts/benchmarks/summarize_public_matrix.py

# Full benchmark matrix
benchmark-full:
	./scripts/benchmarks/run_public_matrix.sh full
	python3 ./scripts/benchmarks/summarize_public_matrix.py

# Stress benchmark (all repos including very large ones)
benchmark-stress:
	./scripts/benchmarks/run_public_matrix.sh stress
	python3 ./scripts/benchmarks/summarize_public_matrix.py

# Just regenerate the summary from existing artifacts
benchmark-summary:
	python3 ./scripts/benchmarks/summarize_public_matrix.py
