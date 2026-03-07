#!/usr/bin/env bash
# Release gate — verifies all quality bars before release.
# Exit on first failure.
set -euo pipefail

echo "=== Hamlet Release Gate ==="
echo

echo "1. go vet"
go vet ./cmd/... ./internal/...
echo "   PASS"
echo

echo "2. go build"
go build -o /dev/null ./cmd/hamlet/
echo "   PASS"
echo

echo "3. Unit and integration tests"
go test ./internal/... -count=1 -timeout 120s
echo "   PASS"
echo

echo "4. Testdata suite (golden, determinism, schema, adversarial, E2E, CLI)"
go test ./internal/testdata/ -count=1 -timeout 120s
echo "   PASS"
echo

echo "5. Golden file verification"
go test ./internal/testdata/ -run TestGolden -count=1
echo "   PASS"
echo

echo "6. Determinism verification (10 iterations)"
go test ./internal/testdata/ -run TestDeterminism -count=1
echo "   PASS"
echo

echo "=== All gates passed ==="
