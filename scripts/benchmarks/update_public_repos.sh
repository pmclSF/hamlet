#!/usr/bin/env bash
# Update already-cloned benchmark repos to latest.
#
# Usage:
#   ./scripts/benchmarks/update_public_repos.sh              # update all
#   ./scripts/benchmarks/update_public_repos.sh --tier smoke  # smoke-tier only
#   ./scripts/benchmarks/update_public_repos.sh --id express  # one repo
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPOS_DIR="$ROOT_DIR/benchmarks/repos"

TIER_FILTER=""
ID_FILTER=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        --tier)  TIER_FILTER="$2"; shift 2 ;;
        --id)    ID_FILTER="$2"; shift 2 ;;
        -h|--help) head -6 "$0" | tail -4; exit 0 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

MANIFEST="$ROOT_DIR/benchmarks/public-repos.yaml"

parse_manifest() {
    python3 -c "
import yaml, sys
with open('$MANIFEST') as f:
    data = yaml.safe_load(f)
for repo in data.get('repos', []):
    print(f\"{repo['id']}|{repo.get('tier','full')}\")
"
}

echo "=== Terrain Benchmark: Update Public Repos ==="

UPDATED=0
MISSING=0
FAILED=0

while IFS='|' read -r id tier; do
    if [[ -n "$TIER_FILTER" && "$tier" != "$TIER_FILTER" ]]; then
        continue
    fi
    if [[ -n "$ID_FILTER" && "$id" != "$ID_FILTER" ]]; then
        continue
    fi

    repo_dir="$REPOS_DIR/$id"
    if [[ ! -d "$repo_dir/.git" ]]; then
        echo "[$id] Not cloned — skipping (run fetch first)"
        MISSING=$((MISSING + 1))
        continue
    fi

    echo -n "[$id] Updating... "
    if git -C "$repo_dir" pull --ff-only 2>&1 | tail -1 | sed 's/^/  /'; then
        UPDATED=$((UPDATED + 1))
    else
        echo "  FAILED" >&2
        FAILED=$((FAILED + 1))
    fi
done < <(parse_manifest)

echo ""
echo "=== Update complete ==="
echo "  Updated: $UPDATED  Missing: $MISSING  Failed: $FAILED"
