#!/usr/bin/env bash
# Fetch public repos for benchmarking.
#
# Usage:
#   ./scripts/benchmarks/fetch_public_repos.sh              # fetch all
#   ./scripts/benchmarks/fetch_public_repos.sh --tier smoke  # fetch smoke-tier only
#   ./scripts/benchmarks/fetch_public_repos.sh --id express  # fetch one repo
#   ./scripts/benchmarks/fetch_public_repos.sh --full-clone  # full clone (not shallow)
#   ./scripts/benchmarks/fetch_public_repos.sh --force       # re-clone even if exists
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
MANIFEST="$ROOT_DIR/benchmarks/public-repos.yaml"
REPOS_DIR="$ROOT_DIR/benchmarks/repos"

# Defaults
TIER_FILTER=""
ID_FILTER=""
FULL_CLONE=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --tier)    TIER_FILTER="$2"; shift 2 ;;
        --id)      ID_FILTER="$2"; shift 2 ;;
        --full-clone) FULL_CLONE=true; shift ;;
        --force)   FORCE=true; shift ;;
        -h|--help)
            head -8 "$0" | tail -6
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

if ! command -v python3 &>/dev/null; then
    echo "Error: python3 is required to parse the manifest." >&2
    exit 1
fi

mkdir -p "$REPOS_DIR"

# Parse YAML manifest with a small Python helper.
# Outputs: id|url|branch|tier|clone_strategy — one line per repo.
parse_manifest() {
    python3 -c "
import yaml, sys

with open('$MANIFEST') as f:
    data = yaml.safe_load(f)

for repo in data.get('repos', []):
    rid = repo['id']
    url = repo['url']
    branch = repo.get('branch', 'main')
    tier = repo.get('tier', 'full')
    clone = repo.get('clone', 'shallow')
    print(f'{rid}|{url}|{branch}|{tier}|{clone}')
"
}

TOTAL=0
SUCCESS=0
SKIPPED=0
FAILED=0

echo "=== Hamlet Benchmark: Fetch Public Repos ==="
echo "  Repos dir: $REPOS_DIR"
echo "  Tier filter: ${TIER_FILTER:-all}"
echo "  ID filter: ${ID_FILTER:-all}"
echo ""

while IFS='|' read -r id url branch tier clone_strategy; do
    # Apply filters.
    if [[ -n "$TIER_FILTER" && "$tier" != "$TIER_FILTER" ]]; then
        continue
    fi
    if [[ -n "$ID_FILTER" && "$id" != "$ID_FILTER" ]]; then
        continue
    fi

    TOTAL=$((TOTAL + 1))
    repo_dir="$REPOS_DIR/$id"

    # Skip if already exists (unless --force).
    if [[ -d "$repo_dir/.git" && "$FORCE" != "true" ]]; then
        echo "[$id] Already cloned — skipping (use --force to re-clone)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Remove if forcing.
    if [[ -d "$repo_dir" && "$FORCE" == "true" ]]; then
        echo "[$id] Removing existing clone (--force)"
        rm -rf "$repo_dir"
    fi

    echo "[$id] Cloning $url (branch: $branch)..."

    clone_args=("git" "clone")
    if [[ "$FULL_CLONE" != "true" && "$clone_strategy" == "shallow" ]]; then
        clone_args+=("--depth" "1")
    fi
    clone_args+=("--branch" "$branch" "--single-branch" "$url" "$repo_dir")

    if "${clone_args[@]}" 2>&1 | sed 's/^/  /'; then
        # Show size.
        size=$(du -sh "$repo_dir" 2>/dev/null | cut -f1)
        echo "[$id] Done ($size)"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "[$id] FAILED to clone — continuing" >&2
        FAILED=$((FAILED + 1))
    fi
    echo ""
done < <(parse_manifest)

echo "=== Fetch complete ==="
echo "  Total: $TOTAL  Success: $SUCCESS  Skipped: $SKIPPED  Failed: $FAILED"

if [[ $FAILED -gt 0 ]]; then
    exit 1
fi
