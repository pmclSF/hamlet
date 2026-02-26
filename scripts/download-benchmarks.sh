#!/usr/bin/env bash
# Download benchmark repos for Hamlet round-trip testing.
#
# Usage:
#   ./scripts/download-benchmarks.sh            # download all categories
#   ./scripts/download-benchmarks.sh vitest      # download one category
#   ./scripts/download-benchmarks.sh --dry-run   # preview only
#
# Idempotent: skips repos that already exist on disk.
# Removes .git/ dirs after clone to save ~40% disk space.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BENCH_DIR="$ROOT_DIR/benchmarks"

DRY_RUN=false
CATEGORY=""

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    *) CATEGORY="$arg" ;;
  esac
done

# Clone a repo (shallow, depth=1) into the given directory.
# Usage: clone_repo <url> <dest> [--branch <tag>]
clone_repo() {
  local url="$1"
  local dest="$2"
  local branch_args=()

  if [[ "${3:-}" == "--branch" ]]; then
    branch_args=(--branch "$4")
  fi

  if [[ -d "$dest" ]]; then
    echo "  SKIP  $dest (already exists)"
    return
  fi

  if $DRY_RUN; then
    echo "  WOULD CLONE  $url -> $dest ${branch_args[*]:-}"
    return
  fi

  echo "  CLONE  $url -> $dest"
  git clone --depth 1 "${branch_args[@]}" "$url" "$dest" 2>&1 | sed 's/^/    /'

  # Remove .git dir to save disk space
  rm -rf "$dest/.git"
  echo "    Removed .git/ (saved space)"
}

# ── Vitest repos ──────────────────────────────────────────────────────
download_vitest() {
  echo ""
  echo "=== Vitest repos ==="
  local dir="$BENCH_DIR/vitest"
  mkdir -p "$dir"

  clone_repo "https://github.com/vuejs/pinia.git" "$dir/pinia"
  clone_repo "https://github.com/vitejs/vite.git" "$dir/vite"
  clone_repo "https://github.com/unocss/unocss.git" "$dir/unocss"
}

# ── Mocha repos (additional) ─────────────────────────────────────────
download_mocha() {
  echo ""
  echo "=== Mocha repos (additional) ==="
  local dir="$BENCH_DIR/mocha-jasmine-etc"
  mkdir -p "$dir"

  clone_repo "https://github.com/chaijs/chai.git" "$dir/chai"
  clone_repo "https://github.com/socketio/socket.io.git" "$dir/socket.io"
  clone_repo "https://github.com/Automattic/mongoose.git" "$dir/mongoose"
}

# ── Python/Pytest repos (additional) ─────────────────────────────────
download_python() {
  echo ""
  echo "=== Python/Pytest repos (additional) ==="
  local dir="$BENCH_DIR/python"
  mkdir -p "$dir"

  clone_repo "https://github.com/pallets/flask.git" "$dir/flask"
  clone_repo "https://github.com/psf/requests.git" "$dir/requests"
  clone_repo "https://github.com/encode/httpx.git" "$dir/httpx"
}

# ── Java/JUnit4 repos ────────────────────────────────────────────────
download_java() {
  echo ""
  echo "=== Java/JUnit4 repos ==="
  local dir="$BENCH_DIR/java"
  mkdir -p "$dir"

  # Mockito v4.11.0 — last major JUnit4-era release before JUnit5 migration
  clone_repo "https://github.com/mockito/mockito.git" "$dir/mockito" --branch "v4.11.0"
  clone_repo "https://github.com/google/guava.git" "$dir/guava"
}

# ── Main ──────────────────────────────────────────────────────────────
echo "Hamlet Benchmark Repo Downloader"
echo "Benchmark dir: $BENCH_DIR"
if $DRY_RUN; then
  echo "Mode: DRY RUN (no changes)"
fi

if [[ -z "$CATEGORY" ]]; then
  download_vitest
  download_mocha
  download_python
  download_java
else
  case "$CATEGORY" in
    vitest)  download_vitest ;;
    mocha)   download_mocha ;;
    python)  download_python ;;
    java)    download_java ;;
    *)
      echo "Unknown category: $CATEGORY"
      echo "Available: vitest, mocha, python, java"
      exit 1
      ;;
  esac
fi

echo ""
echo "Done."
