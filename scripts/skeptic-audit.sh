#!/usr/bin/env bash
# Docs-and-surface audit: keeps the shipped CLI surface, the docs, and the
# tracked tree consistent.
#
# All greps run against tracked content only (git grep / git ls-files);
# gitignored local files are never scanned.
#
#   1. no references to the dead terrain[.]dev domain outside the Makefile
#      guard that names it
#   2. no roadmap placeholder strings in cmd/ or internal/ Go string
#      literals
#   3. `terrain --help` lists every dispatched top-level command
#      (dev-gated commands excepted)
#   4. no compiled binaries tracked at the repository root
#   5. [Unreleased] in CHANGELOG.md is non-empty when commits exist past
#      the latest release tag
#   6. a bare `terrain` run at the repo root maps no gitignored content
#
# Prints one PASS/FAIL line per check; exits non-zero if any check fails.

set -u

cd "$(git rev-parse --show-toplevel)" || exit 1

status=0

pass() { printf 'PASS  %s\n' "$1"; }
fail() {
	printf 'FAIL  %s\n' "$1"
	status=1
}
warn() { printf 'WARN  %s\n' "$1"; }
detail() { printf '%s\n' "$1" | sed 's/^/      /'; }

# ── 1. dead domain ─────────────────────────────────────────────────────
# The Makefile's no-dead-domains guard necessarily names the domain in its
# own pattern and message, so the Makefile is the one allowed location.
# The bracketed character class keeps this script from matching itself.
hits="$(git grep -nIE 'terrain[.]dev' -- ':!Makefile' 2>/dev/null || true)"
if [ -n "$hits" ]; then
	fail "check 1: terrain[.]dev referenced in tracked files"
	detail "$hits"
else
	pass "check 1: no terrain[.]dev references in tracked files"
fi

# ── 2. roadmap placeholders ────────────────────────────────────────────
# User-facing strings must describe what ships today. Whole-line comments
# are excluded (handled in review); test files are not user-facing.
hits="$(git grep -nIiE 'deferred to a future release|reserved for a future release|coming soon' \
	-- 'cmd/*.go' 'internal/*.go' ':!*_test.go' 2>/dev/null |
	grep -vE '^[^:]+:[0-9]+:[[:space:]]*//' || true)"
if [ -n "$hits" ]; then
	fail "check 2: roadmap placeholder strings in Go sources"
	detail "$hits"
else
	pass "check 2: no roadmap placeholder strings in cmd/ or internal/"
fi

# ── 3. help coverage ───────────────────────────────────────────────────
# Every top-level command the dispatcher accepts must appear in
# `terrain --help`, except commands gated behind TERRAIN_DEV. The expected
# list is derived from the knownCommands table in cmd/terrain/main.go,
# which the dispatcher keeps in sync with its switch.
bin="${TMPDIR:-/tmp}/terrain-audit"
if go build -o "$bin" ./cmd/terrain; then
	help_out="$("$bin" --help 2>&1)"
	cmds="$(awk '/^var knownCommands = /,/^}/' cmd/terrain/main.go |
		grep -o '"[^"]*"' | tr -d '"')"
	if [ -z "$cmds" ]; then
		fail "check 3: could not derive the command list from cmd/terrain/main.go"
	else
		missing=""
		for c in $cmds; do
			case "$c" in
			help | --help | -h) continue ;; # help aliases, not commands
			webhook) continue ;;            # opt-in: gated behind TERRAIN_DEV
			esac
			printf '%s\n' "$help_out" | grep -qw -- "$c" || missing="$missing $c"
		done
		if [ -n "$missing" ]; then
			fail "check 3: --help is missing dispatched commands:$missing"
		else
			pass "check 3: --help covers every dispatched top-level command"
		fi
	fi
	# Floor: these two must be present even if list derivation changes.
	for c in scaffold inject; do
		printf '%s\n' "$help_out" | grep -qw -- "$c" ||
			fail "check 3: --help is missing $c"
	done
else
	bin=""
	fail "check 3: go build ./cmd/terrain failed"
fi

# ── 4. no tracked binaries ─────────────────────────────────────────────
# Compiled outputs must never be committed. Exact names cover the build
# targets; the mode+contents scan catches any other executable blob at
# the repository root.
named="$(git ls-files -- terrain terrain-bench terrain-corpus \
	terrain-mechanism-recall terrain-voice-lint)"
blobs=""
while IFS=$'\t' read -r meta path; do
	[ -n "$path" ] || continue
	[ "${meta%% *}" = "100755" ] || continue
	[ -f "$path" ] || continue
	if head -c 4096 "$path" | od -An -tx1 | grep -q ' 00'; then
		blobs="$blobs $path"
	fi
done < <(git ls-files -s -- . ':!*/*')
if [ -n "$named$blobs" ]; then
	fail "check 4: compiled binaries tracked at the repository root"
	detail "$(printf '%s\n%s\n' "$named" "$blobs" | grep '[^[:space:]]')"
else
	pass "check 4: no compiled binaries tracked at the repository root"
fi

# ── 5. changelog discipline ────────────────────────────────────────────
if desc="$(git describe --tags --match 'v*' --long 2>/dev/null)"; then
	ahead="${desc%-g*}"
	ahead="${ahead##*-}"
	case "$ahead" in
	'' | *[!0-9]*)
		warn "check 5: could not parse git describe output ($desc) — skipped"
		;;
	0)
		pass "check 5: no commits past the latest release tag"
		;;
	*)
		body="$(awk '/^## \[Unreleased\]/{f=1; next} /^## \[/{f=0} f' \
			CHANGELOG.md | grep -c '[^[:space:]]')"
		if [ "$body" -gt 0 ]; then
			pass "check 5: [Unreleased] has content ($ahead commit(s) past the latest tag)"
		else
			fail "check 5: $ahead commit(s) past the latest tag but [Unreleased] is empty"
		fi
		;;
	esac
else
	warn "check 5: no v* release tag reachable (shallow clone?) — skipped"
fi

# ── 6. clean discovery output ──────────────────────────────────────────
# A bare `terrain` run at the repo root must not map gitignored local
# content (tier-4/ is gitignored here; see .gitignore).
if [ -n "$bin" ]; then
	run_out="$("$bin" 2>&1)" && run_rc=0 || run_rc=$?
	if [ -z "$run_out" ] ||
		{ [ "$run_rc" -ne 0 ] && ! printf '%s\n' "$run_out" | grep -q 'MAPPED'; }; then
		warn "check 6: no discovery report produced (exit $run_rc) — skipped"
	else
		ignored="$(printf '%s\n' "$run_out" | grep -n 'tier-4/' || true)"
		if [ -n "$ignored" ]; then
			fail "check 6: discovery output includes gitignored paths"
			detail "$ignored"
		else
			pass "check 6: discovery output includes no gitignored paths"
		fi
	fi
else
	warn "check 6: skipped (build failed)"
fi

if [ "$status" -ne 0 ]; then
	printf 'audit: FAIL\n'
else
	printf 'audit: PASS\n'
fi
exit "$status"
