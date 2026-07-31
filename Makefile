# Terrain — signal-first test intelligence platform

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
GO_OWNED_PKGS := ./cmd/... ./internal/...

.PHONY: build test lint clean demo benchmark-fetch benchmark-smoke benchmark-full benchmark-stress benchmark-summary benchmark-convert install docs-linkcheck \
       test-golden test-determinism test-schema test-adversarial test-e2e test-cli test-bench golden-update pr-gate release-gate \
       sbom sbom-cyclonedx sbom-spdx release-dry-run go-release-verify js-release-verify extension-verify release-verify \
       docs-gen docs-verify calibrate bench-baseline bench-gate memory-bench truth-verify voice-lint no-dead-domains offline-check

# Build the CLI binary
build:
	go build -ldflags "$(LDFLAGS)" -o terrain ./cmd/terrain

# Install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/terrain

# Run all Go tests
test:
	go test $(GO_OWNED_PKGS)

# Run tests with verbose output
test-v:
	go test -v $(GO_OWNED_PKGS)

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out $(GO_OWNED_PKGS)
	go tool cover -func=coverage.out

# Build check (compile only, no binary output)
check:
	go vet $(GO_OWNED_PKGS)
	go build ./cmd/terrain

# Clean build artifacts
clean:
	rm -f terrain coverage.out terrain.cdx.json terrain.spdx.json

# Run demo: analyze the current repository
demo:
	@echo "=== Terrain Demo ==="
	@echo ""
	@echo "--- terrain analyze ---"
	go run ./cmd/terrain analyze
	@echo ""
	@echo "--- terrain summary ---"
	go run ./cmd/terrain summary
	@echo ""
	@echo "--- terrain metrics ---"
	go run ./cmd/terrain metrics

# ── Verification Layer Targets ──────────────────────────────

# Run golden tests only (fast)
test-golden:
	go test ./internal/testdata/ -run 'Golden' -count=1

# Run determinism tests only
test-determinism:
	go test ./internal/testdata/ -run 'Determinism' -count=1

# Run schema tests only
test-schema:
	go test ./internal/testdata/ -run 'Schema' -count=1

# Run adversarial tests only
test-adversarial:
	go test ./internal/testdata/ -run 'Adversarial' -count=1

# Run E2E scenario tests only
test-e2e:
	go test ./internal/testdata/ -run 'E2E' -count=1

# Run CLI regression tests only
test-cli:
	go test ./internal/testdata/ -run 'CLI' -count=1 -timeout 120s

# Run benchmarks
test-bench:
	go test -bench . -benchmem ./internal/testdata/ -run '^$$'

# Update golden files (review changes in git diff before committing)
golden-update:
	go test ./internal/testdata/ -run 'Golden' -update

# Run the offline-guarantee assertions locally (network up; the CI job
# runs the same script inside an isolated network namespace on ubuntu)
offline-check: build
	work=$$(mktemp -d) && \
	cp tests/fixtures/offline/*.py "$$work/" && \
	bash tests/fixtures/offline/check.sh "$(CURDIR)/terrain" "$$work" "$(CURDIR)/tests/fixtures/offline" && \
	rm -rf "$$work"

# ── SBOM Generation ───────────────────────────────────────
# Requires: syft (https://github.com/anchore/syft)
#   brew install syft  OR  go install github.com/anchore/syft/cmd/syft@latest

# Generate CycloneDX SBOM from the built binary
sbom-cyclonedx: build
	syft terrain --output cyclonedx-json=terrain.cdx.json
	@echo "CycloneDX SBOM written to terrain.cdx.json"

# Generate SPDX SBOM from the built binary
sbom-spdx: build
	syft terrain --output spdx-json=terrain.spdx.json
	@echo "SPDX SBOM written to terrain.spdx.json"

# Generate both SBOM formats
sbom: sbom-cyclonedx sbom-spdx

# Dry-run GoReleaser to verify config and preview artifacts
release-dry-run:
	goreleaser release --snapshot --clean --skip=publish,sign

# ── Release Gates ──────────────────────────────────────────

# PR gate: fast checks required on every PR
pr-gate:
	$(MAKE) check
	$(MAKE) no-dead-domains
	$(MAKE) test

# Guard: the terrain.dev domain is not live, so it must never appear
# anywhere in tracked source — a dead link erodes trust. Uses git grep
# (tracked files only).
no-dead-domains:
	@if git grep -n -e 'terrain\.dev' -- ':!Makefile' >/dev/null 2>&1; then \
		echo "dead 'terrain.dev' user-facing reference(s) in tracked source:"; \
		git grep -n -e 'terrain\.dev' -- ':!Makefile'; \
		exit 1; \
	fi
	@echo "no-dead-domains: ok"

# Validate regression suites + recall harnesses for shadow→live flips.
# Loads every YAML under harness/regression-suites/ and
# harness/recall-harnesses/, validates schema, prints a summary.
# Exit non-zero on any load failure.
regression-gate:
	go run ./cmd/internal/terrain-regression-gate

# Release gate: full verification required before release
release-gate: go-release-verify regression-gate

go-release-verify:
	@offenders=$$(gofmt -l .); \
		if [ -n "$$offenders" ]; then \
			echo "::error::gofmt violations:"; echo "$$offenders"; exit 1; \
		fi
	go vet $(GO_OWNED_PKGS)
	go test $(GO_OWNED_PKGS)
	go build ./cmd/terrain
	go test ./cmd/terrain/ -run TestSnapshot -count=1 -v

npm-release-verify:
	npm ci
	npm run release:verify

extension-verify:
	npm --prefix extension/vscode ci
	npm --prefix extension/vscode run compile
	npm --prefix extension/vscode test

# ── Generated documentation ─────────────────────────────────
# `docs-gen` rewrites docs/signals/manifest.json from
# internal/signals.allSignalManifest. `docs-verify` writes to a tempdir
# and diffs against the committed copy so CI fails when a manifest
# change ships without the regenerated docs.
docs-gen:
	go run ./cmd/internal/terrain-docs-gen

docs-verify:
	@scripts/docs-verify.sh
	@docs/scripts/verify_doc_consistency.sh

# `docs-linkcheck` walks docs/ and verifies that every intra-repo
# markdown link resolves to a real file. Skips docs/internal/ and
# docs/legacy/ by default; run with -include-internal to also scan
# them. External links (http/https/mailto) are out of scope.
docs-linkcheck:
	@go run ./cmd/internal/terrain-docs-linkcheck

# `truth-verify` cross-checks docs/release/feature-status.md against
# the canonical signal manifest. Every signal name documented in the
# curated table must reference a real manifest entry; references that
# don't resolve (typo, renamed, removed) fail the build. Orphan stable
# signals (in the manifest, not in the curated doc) print as
# advisory warnings — pass --strict-orphans to fail on them too.
truth-verify:
	@go run ./cmd/internal/terrain-truth-verify

# `voice-lint` enforces voice-and-tone rules: no exclamation-mark
# prose (jarring), no British spellings (mixed-spelling reads as
# under-edited). Scans Go source in the user-visible code paths
# (signals manifest, command package, reporting, changescope). Test
# files are skipped — tests can use any prose without tripping the
# lint.
voice-lint:
	@go run ./cmd/internal/terrain-voice-lint

# ── Recall-regression fixture suite ─────────────────────────
# Runs the engine pipeline against the bundled fixture set and prints
# per-detector recall. Any unmatched expected label fails the build.
calibrate:
	go test -count=1 -v -run TestCalibration ./internal/engine/...

# ── Performance regression gate ─────────────────────────────
# bench-baseline writes a fresh baseline benchmark snapshot. Run on a
# main-branch commit and commit the result.
# bench-gate runs the same benchmarks now and compares against the
# committed baseline; fails if any benchmark regressed >10%.
bench-baseline:
	go test -run '^$$' -bench 'BenchmarkRunPipeline|BenchmarkSignalDetection|BenchmarkBuildImportGraph|BenchmarkRiskScore|BenchmarkExtractTestCases' \
		-count=5 ./internal/engine ./internal/analysis ./internal/scoring ./internal/testcase \
		> benchmarks/baseline.txt
	@echo "Wrote benchmarks/baseline.txt"

# `memory-bench` runs the memory ceiling + leak-detection tests
# (TestMemoryCeiling_*, TestMemoryNoLeak_*). Skipped in the default
# `go test ./...` loop because they're slow (force GC + run analysis
# at scale). Set TERRAIN_MEMORY_BENCH=1 inline; this target does it
# for you.
memory-bench:
	@TERRAIN_MEMORY_BENCH=1 go test -v -count=1 -run 'TestMemory' ./internal/analysis/...

bench-gate:
	@tmp=$$(mktemp) ; \
	go test -run '^$$' -bench 'BenchmarkRunPipeline|BenchmarkSignalDetection|BenchmarkBuildImportGraph|BenchmarkRiskScore|BenchmarkExtractTestCases' \
		-count=5 ./internal/engine ./internal/analysis ./internal/scoring ./internal/testcase > $$tmp ; \
	go run ./cmd/internal/terrain-bench-gate --base benchmarks/baseline.txt --head $$tmp --threshold 10 ; \
	rc=$$? ; \
	rm -f $$tmp ; \
	exit $$rc

release-verify:
	$(MAKE) go-release-verify
	$(MAKE) npm-release-verify
	$(MAKE) extension-verify

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

# Compare current Go converters against the legacy JS runtime floor.
benchmark-convert:
	go run ./cmd/internal/terrain-convert-bench
