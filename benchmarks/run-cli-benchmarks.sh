#!/usr/bin/env bash
# Run Terrain CLI benchmarks against configured repositories.
#
# Usage:
#   ./benchmarks/run-cli-benchmarks.sh                     # run all
#   ./benchmarks/run-cli-benchmarks.sh --repo terrain       # single repo
#   ./benchmarks/run-cli-benchmarks.sh --command analyze   # single command
#   ./benchmarks/run-cli-benchmarks.sh --discover /path    # auto-discover repos
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

exec go run ./cmd/terrain-bench/ "$@"
