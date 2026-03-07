# Hamlet — observability and intelligence for test suites

.PHONY: build test lint clean demo

# Build the CLI binary
build:
	go build -o hamlet ./cmd/hamlet

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
