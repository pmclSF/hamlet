#!/usr/bin/env bash
# Offline-guarantee assertions: terrain's discover / analyze / gate / fix
# path must complete — and do real work — with zero network access.
#
# Usage: check.sh <terrain-binary> <workdir> <fixture-src-dir>
#
#   <terrain-binary>   absolute path to a built terrain binary
#   <workdir>          scratch directory pre-seeded with this fixture's .py
#                      files; all output lands here
#   <fixture-src-dir>  absolute path to the pristine fixture (used to prove
#                      the fix dry-run wrote nothing)
#
# In CI this script runs inside `sudo unshare -n` — a network namespace with
# no interfaces, no routes, and no DNS — so any attempted egress fails the
# command and the job. Locally (`make offline-check`) it runs with the
# network up and validates the fixture + assertions themselves.
#
# `terrain serve` is deliberately not exercised: it binds loopback by design.
set -euo pipefail

bin="$1"
work="$2"
src="$3"
cd "$work"

echo '--- 1/4 discovery report (bare terrain) ---'
"$bin" | tee discover.out
grep -q '1 prompt' discover.out
grep -q '1 schema' discover.out
grep -q 'drift' discover.out

echo '--- 2/4 terrain analyze finds the seeded drift ---'
"$bin" analyze --root . --json > analyze.json
jq -e '[.signals[] | select(.type == "aiPromptSchemaDrift" and .severity == "high")] | length == 1' analyze.json > /dev/null

echo '--- 3/4 terrain test --fail-on high produces its verdict ---'
rc=0
"$bin" test --fail-on high --junit junit.xml > test.out 2>&1 || rc=$?
cat test.out
# 6 is the documented severity-gate exit (see `terrain test --help`).
# Any other status means the pipeline crashed instead of gating.
if [ "$rc" -ne 6 ]; then
	echo "expected severity-gate exit 6, got $rc" >&2
	exit 1
fi
[ -s junit.xml ]
grep -q 'prompt-schema-drift' junit.xml

echo '--- 4/4 terrain fix dry-run proposes, writes nothing ---'
"$bin" fix | tee fix.out
grep -qi 'validated fix' fix.out
grep -q 'nothing written' fix.out
cmp "$src/prompt.py" prompt.py

echo 'offline guarantee holds: discover, analyze, gate, and fix all completed'
