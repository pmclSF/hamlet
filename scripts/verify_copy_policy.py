#!/usr/bin/env python3
"""Enforce release-copy policy for active product surfaces.

Policy:
- No numbered release-line branding in active Markdown copy.
- No numbered release-line branding in CLI help output.
- No numbered release-line branding in benchmark-safe JSON outputs.

Intentional exclusions:
- CHANGELOG.md (historical release chronology)
- test/ tests/ fixtures/ (non-customer fixture content)
- technical protocol/dependency versions are handled outside this check
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

DISALLOWED = re.compile(r"\b[Vv](?:2|3)(?:\.[0-9xX]+){0,2}\b")
ANALYSIS_VERSION_DISALLOWED = re.compile(r'"analysisVersion"\s*:\s*"v[0-9]')

EXCLUDED_TOP_LEVEL_PREFIXES = (
    "benchmarks/",
    "terrain/",
    "node_modules/",
    "dist/",
    "artifacts/",
    "reports/",
    "test/",
    "tests/",
    "fixtures/",
)

EXCLUDED_FILES = {"CHANGELOG.md"}


def should_scan_markdown(rel: str) -> bool:
    if rel in EXCLUDED_FILES:
        return False
    if any(rel.startswith(prefix) for prefix in EXCLUDED_TOP_LEVEL_PREFIXES):
        return False
    parts = rel.split("/")
    if "node_modules" in parts or ".git" in parts:
        return False
    return True


def scan_markdown() -> list[str]:
    violations: list[str] = []
    for path in ROOT.rglob("*.md"):
        rel = path.relative_to(ROOT).as_posix()
        if not should_scan_markdown(rel):
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        for idx, line in enumerate(text.splitlines(), start=1):
            if DISALLOWED.search(line):
                violations.append(f"{rel}:{idx}: {line.strip()}")
    return violations


def run(cmd: list[str], cwd: Path, env: dict[str, str] | None = None) -> str:
    proc = subprocess.run(cmd, cwd=cwd, text=True, capture_output=True, env=env)
    if proc.returncode != 0:
        raise RuntimeError(
            f"command failed: {' '.join(cmd)}\nstdout:\n{proc.stdout}\nstderr:\n{proc.stderr}"
        )
    return proc.stdout


def verify_cli_and_json() -> list[str]:
    violations: list[str] = []
    with tempfile.TemporaryDirectory(prefix="terrain-copy-policy-") as td:
        temp = Path(td)
        go_cache = temp / "go-build-cache"
        go_tmp = temp / "go-tmp"
        go_cache.mkdir(parents=True, exist_ok=True)
        go_tmp.mkdir(parents=True, exist_ok=True)
        go_env = os.environ.copy()
        go_env["GOCACHE"] = str(go_cache)
        go_env["GOTMPDIR"] = str(go_tmp)

        binary = temp / "terrain"
        run(["go", "build", "-o", str(binary), "./cmd/terrain"], ROOT, env=go_env)

        help_output = run([str(binary), "--help"], ROOT)
        if DISALLOWED.search(help_output):
            violations.append("CLI help output contains disallowed numbered release-line branding.")

        repo = temp / "repo"
        (repo / "src").mkdir(parents=True, exist_ok=True)
        (repo / "tests").mkdir(parents=True, exist_ok=True)
        (repo / "src" / "math.js").write_text(
            "export function add(a, b) { return a + b; }\n", encoding="utf-8"
        )
        (repo / "tests" / "math.test.js").write_text(
            "describe('math', () => { it('adds', () => { expect(1 + 2).toBe(3); }); });\n",
            encoding="utf-8",
        )

        metrics_json = run([str(binary), "metrics", "--root", str(repo), "--json"], ROOT)
        benchmark_json = run([str(binary), "export", "benchmark", "--root", str(repo)], ROOT)

        for name, payload in (("metrics", metrics_json), ("export benchmark", benchmark_json)):
            if DISALLOWED.search(payload):
                violations.append(f"{name} JSON output contains disallowed numbered release-line branding.")
            if ANALYSIS_VERSION_DISALLOWED.search(payload):
                violations.append(
                    f"{name} JSON output contains disallowed analysisVersion prefix (v<digit>)."
                )
            try:
                json.loads(payload)
            except json.JSONDecodeError as exc:
                violations.append(f"{name} output is not valid JSON: {exc}")

    return violations


def main() -> int:
    violations = scan_markdown()
    try:
        violations.extend(verify_cli_and_json())
    except RuntimeError as exc:
        print("copy policy check failed while executing commands:")
        print(exc)
        return 1

    if violations:
        print("copy policy violations found:")
        for v in violations:
            print(f"- {v}")
        return 1

    print("copy policy check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
