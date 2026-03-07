#!/usr/bin/env bash
# Run Hamlet benchmark matrix against public repos.
#
# Usage:
#   ./scripts/benchmarks/run_public_matrix.sh smoke           # smoke tier
#   ./scripts/benchmarks/run_public_matrix.sh full            # full tier
#   ./scripts/benchmarks/run_public_matrix.sh stress          # all tiers
#   ./scripts/benchmarks/run_public_matrix.sh --id express    # one repo
#   ./scripts/benchmarks/run_public_matrix.sh full --skip-determinism
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
MANIFEST="$ROOT_DIR/benchmarks/public-repos.yaml"
REPOS_DIR="$ROOT_DIR/benchmarks/repos"
ARTIFACTS_DIR="$ROOT_DIR/artifacts/public-benchmarks"
HAMLET_BIN="$ROOT_DIR/hamlet-bench"

# Defaults
MODE="${1:-smoke}"
shift || true
ID_FILTER=""
SKIP_DETERMINISM=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --id) ID_FILTER="$2"; shift 2 ;;
        --skip-determinism) SKIP_DETERMINISM=true; shift ;;
        -h|--help) head -8 "$0" | tail -6; exit 0 ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

# Determine which tiers to include.
case "$MODE" in
    smoke)  TIERS="smoke" ;;
    full)   TIERS="smoke full" ;;
    stress) TIERS="smoke full stress" ;;
    *)
        echo "Unknown mode: $MODE (use smoke, full, or stress)" >&2
        exit 1
        ;;
esac

# ── Step 1: Build Hamlet ─────────────────────────────────────

echo "=== Hamlet Public Benchmark Matrix ==="
echo "  Mode: $MODE"
echo "  Tiers: $TIERS"
echo ""

echo "Building Hamlet..."
if ! go build -o "$HAMLET_BIN" ./cmd/hamlet/ 2>&1; then
    echo "FATAL: Build failed" >&2
    exit 1
fi
echo "  Built: $HAMLET_BIN"
echo ""

# ── Step 2: Parse manifest ───────────────────────────────────

parse_manifest() {
    python3 -c "
import yaml, sys
with open('$MANIFEST') as f:
    data = yaml.safe_load(f)
for repo in data.get('repos', []):
    print(f\"{repo['id']}|{repo.get('tier','full')}|{repo.get('branch','main')}\")
"
}

# ── Step 3: Run matrix ───────────────────────────────────────

TOTAL=0
PASS=0
FAIL=0
SKIP=0

# Core commands to run on every repo.
COMMANDS=(
    "analyze_json:analyze --json"
    "analyze_text:analyze"
    "summary:summary"
    "posture:posture"
    "metrics_json:metrics --json"
    "export:export benchmark"
)

run_command() {
    local repo_id="$1"
    local cmd_name="$2"
    local cmd_args="$3"
    local repo_dir="$4"
    local out_dir="$5"

    local stdout_file="$out_dir/${cmd_name}.stdout"
    local stderr_file="$out_dir/${cmd_name}.stderr"
    local meta_file="$out_dir/${cmd_name}.meta"

    local start_ns
    start_ns=$(python3 -c "import time; print(int(time.time()*1000))")

    local exit_code=0
    # shellcheck disable=SC2086
    "$HAMLET_BIN" $cmd_args --root "$repo_dir" >"$stdout_file" 2>"$stderr_file" || exit_code=$?

    local end_ns
    end_ns=$(python3 -c "import time; print(int(time.time()*1000))")
    local duration_ms=$((end_ns - start_ns))

    # Write metadata.
    cat > "$meta_file" <<EOF
repo_id: $repo_id
command: $cmd_name
args: $cmd_args
exit_code: $exit_code
duration_ms: $duration_ms
timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)
EOF

    if [[ $exit_code -eq 0 ]]; then
        printf "    %-20s OK  (%d ms)\n" "$cmd_name" "$duration_ms"
    else
        printf "    %-20s FAIL (exit %d, %d ms)\n" "$cmd_name" "$exit_code" "$duration_ms"
    fi

    return $exit_code
}

# Determinism check: run analyze --json twice, compare with timestamps stripped.
run_determinism_check() {
    local repo_id="$1"
    local repo_dir="$2"
    local out_dir="$3"

    local run1="$out_dir/determinism_run1.json"
    local run2="$out_dir/determinism_run2.json"
    local result_file="$out_dir/determinism.meta"

    "$HAMLET_BIN" analyze --json --root "$repo_dir" >"$run1" 2>/dev/null || true
    "$HAMLET_BIN" analyze --json --root "$repo_dir" >"$run2" 2>/dev/null || true

    # Strip timestamps and generatedAt fields for comparison.
    local norm1 norm2
    norm1=$(python3 -c "
import json, sys
with open('$run1') as f:
    d = json.load(f)
def strip_times(obj):
    if isinstance(obj, dict):
        return {k: strip_times(v) for k, v in obj.items()
                if k not in ('generatedAt', 'snapshotTimestamp', 'exportedAt', 'timestamp')}
    if isinstance(obj, list):
        return [strip_times(v) for v in obj]
    return obj
print(json.dumps(strip_times(d), sort_keys=True))
" 2>/dev/null || echo "PARSE_ERROR")

    norm2=$(python3 -c "
import json, sys
with open('$run2') as f:
    d = json.load(f)
def strip_times(obj):
    if isinstance(obj, dict):
        return {k: strip_times(v) for k, v in obj.items()
                if k not in ('generatedAt', 'snapshotTimestamp', 'exportedAt', 'timestamp')}
    if isinstance(obj, list):
        return [strip_times(v) for v in obj]
    return obj
print(json.dumps(strip_times(d), sort_keys=True))
" 2>/dev/null || echo "PARSE_ERROR")

    if [[ "$norm1" == "$norm2" && "$norm1" != "PARSE_ERROR" ]]; then
        echo "determinism: pass" > "$result_file"
        printf "    %-20s OK\n" "determinism"
    else
        echo "determinism: fail" > "$result_file"
        printf "    %-20s FAIL (outputs differ)\n" "determinism"
    fi
}

# Check expectations.
check_expectations() {
    local repo_id="$1"
    local out_dir="$2"
    local expect_file="$ROOT_DIR/benchmarks/expectations/${repo_id}.yaml"
    local result_file="$out_dir/expectations.meta"

    if [[ ! -f "$expect_file" ]]; then
        echo "expectations: skipped (no file)" > "$result_file"
        return
    fi

    local analyze_json="$out_dir/analyze_json.stdout"
    if [[ ! -f "$analyze_json" ]]; then
        echo "expectations: skipped (no analyze output)" > "$result_file"
        return
    fi

    python3 -c "
import yaml, json, sys

with open('$expect_file') as f:
    expect = yaml.safe_load(f) or {}

with open('$analyze_json') as f:
    try:
        snapshot = json.load(f)
    except:
        print('expectations: fail (invalid JSON)')
        sys.exit(0)

failures = []

# Check minimum test files.
min_tf = expect.get('min_test_files', 0)
actual_tf = len(snapshot.get('testFiles', []))
if actual_tf < min_tf:
    failures.append(f'test files: {actual_tf} < {min_tf}')

# Check minimum code units.
min_cu = expect.get('min_code_units', 0)
actual_cu = len(snapshot.get('codeUnits', []))
if actual_cu < min_cu:
    failures.append(f'code units: {actual_cu} < {min_cu}')

# Check that posture dimensions exist.
if expect.get('require_posture', False):
    meas = snapshot.get('measurements')
    if not meas or not meas.get('posture'):
        failures.append('posture dimensions missing')

# Check that analyze must succeed (checked by caller via exit code).

if failures:
    print('expectations: fail')
    for f in failures:
        print(f'  - {f}')
else:
    print('expectations: pass')
    print(f'  test_files: {actual_tf}')
    print(f'  code_units: {actual_cu}')
" > "$result_file"

    local status
    status=$(head -1 "$result_file")
    if [[ "$status" == *"fail"* ]]; then
        printf "    %-20s FAIL\n" "expectations"
        cat "$result_file" | tail -n +2 | sed 's/^/      /'
    else
        printf "    %-20s OK\n" "expectations"
    fi
}

# ── Main loop ─────────────────────────────────────────────────

while IFS='|' read -r id tier branch; do
    # Filter by tier.
    tier_match=false
    for t in $TIERS; do
        if [[ "$tier" == "$t" ]]; then
            tier_match=true
            break
        fi
    done
    if [[ "$tier_match" != "true" ]]; then
        continue
    fi

    # Filter by ID.
    if [[ -n "$ID_FILTER" && "$id" != "$ID_FILTER" ]]; then
        continue
    fi

    TOTAL=$((TOTAL + 1))
    repo_dir="$REPOS_DIR/$id"

    if [[ ! -d "$repo_dir/.git" ]]; then
        echo "[$id] Not cloned — skipping (run fetch first)"
        SKIP=$((SKIP + 1))
        continue
    fi

    echo "[$id] ($tier tier)"

    # Prepare output directory.
    out_dir="$ARTIFACTS_DIR/$id"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"

    repo_failed=false

    for cmd_spec in "${COMMANDS[@]}"; do
        cmd_name="${cmd_spec%%:*}"
        cmd_args="${cmd_spec#*:}"
        if ! run_command "$id" "$cmd_name" "$cmd_args" "$repo_dir" "$out_dir"; then
            repo_failed=true
        fi
    done

    # Determinism check.
    if [[ "$SKIP_DETERMINISM" != "true" ]]; then
        run_determinism_check "$id" "$repo_dir" "$out_dir"
    fi

    # Expectation check.
    check_expectations "$id" "$out_dir"

    if [[ "$repo_failed" == "true" ]]; then
        FAIL=$((FAIL + 1))
    else
        PASS=$((PASS + 1))
    fi

    echo ""
done < <(parse_manifest)

# ── Cleanup ──────────────────────────────────────────────────

rm -f "$HAMLET_BIN"

echo "=== Benchmark Matrix Complete ==="
echo "  Mode: $MODE"
echo "  Total: $TOTAL  Pass: $PASS  Fail: $FAIL  Skip: $SKIP"
echo "  Artifacts: $ARTIFACTS_DIR"
echo ""

if [[ $FAIL -gt 0 ]]; then
    echo "Some repos had failures. Check artifacts for details."
    exit 1
fi
