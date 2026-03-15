#!/usr/bin/env python3
"""Summarize Terrain public benchmark matrix results.

Reads artifacts from artifacts/public-benchmarks/<repo-id>/ and produces:
  - artifacts/public-benchmarks/summary.md   (human-readable)
  - artifacts/public-benchmarks/summary.json (machine-readable)
  - stdout summary table

Usage:
  python3 scripts/benchmarks/summarize_public_matrix.py
"""

import json
import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
ARTIFACTS = ROOT / "artifacts" / "public-benchmarks"


def parse_meta(path: Path) -> dict:
    """Parse a .meta file into a dict."""
    result = {}
    if not path.exists():
        return result
    for line in path.read_text().splitlines():
        if ": " in line:
            key, _, value = line.partition(": ")
            key = key.strip().lstrip("- ")
            result[key] = value.strip()
    return result


def parse_json_safe(path: Path):
    """Parse a JSON file, returning None on failure."""
    if not path.exists():
        return None
    try:
        return json.loads(path.read_text())
    except (json.JSONDecodeError, OSError):
        return None


def summarize_repo(repo_dir: Path) -> dict:
    """Summarize results for one repo."""
    repo_id = repo_dir.name
    entry = {"id": repo_id, "commands": {}, "overall": "pass"}

    # Process each command's metadata.
    for meta_file in sorted(repo_dir.glob("*.meta")):
        name = meta_file.stem
        meta = parse_meta(meta_file)

        if name == "determinism":
            entry["determinism"] = meta.get("determinism", "unknown")
            # Capture per-command determinism details.
            det_details = {k: v for k, v in meta.items()
                          if k not in ("determinism", "determinism_checks")}
            if det_details:
                entry["determinism_details"] = det_details
            continue

        if name == "expectations":
            status = meta.get("expectations", "unknown")
            entry["expectations"] = status
            if "fail" in status:
                entry["overall"] = "fail"
            continue

        cmd_entry = {
            "exit_code": int(meta.get("exit_code", -1)),
            "duration_ms": int(meta.get("duration_ms", 0)),
        }
        entry["commands"][name] = cmd_entry
        if cmd_entry["exit_code"] != 0:
            entry["overall"] = "fail"

    # Extract stats from analyze JSON output.
    analyze_json = parse_json_safe(repo_dir / "analyze_json.stdout")
    if analyze_json:
        entry["test_files"] = len(analyze_json.get("testFiles", []))
        entry["code_units"] = len(analyze_json.get("codeUnits", []))
        entry["frameworks"] = len(analyze_json.get("frameworks", []))

        # Extract posture dimensions.
        meas = analyze_json.get("measurements")
        if meas and meas.get("posture"):
            entry["posture_dimensions"] = len(meas["posture"])
            entry["posture_bands"] = {
                p["dimension"]: p["band"] for p in meas["posture"]
            }

    # Extract portfolio data.
    portfolio_json = parse_json_safe(repo_dir / "portfolio_json.stdout")
    if portfolio_json:
        agg = portfolio_json.get("aggregates", {})
        entry["portfolio_posture"] = agg.get("portfolioPostureBand", "unknown")
        entry["portfolio_assets"] = agg.get("totalAssets", len(portfolio_json.get("assets", [])))

    # Extract migration readiness data.
    migration_json = parse_json_safe(repo_dir / "migration_readiness_json.stdout")
    if migration_json:
        entry["migration_readiness"] = migration_json.get("readinessLevel", "unknown")
        entry["migration_blockers"] = migration_json.get("totalBlockers", 0)
        fws = migration_json.get("frameworks", [])
        entry["detected_frameworks"] = [f.get("name", "?") for f in fws]

    # Total duration across all commands.
    entry["total_duration_ms"] = sum(
        c["duration_ms"] for c in entry["commands"].values()
    )

    # Determinism.
    if "determinism" not in entry:
        entry["determinism"] = "skipped"

    if entry.get("determinism") == "fail":
        entry["overall"] = "degraded"

    return entry


def format_duration(ms: int) -> str:
    """Format milliseconds as human-readable duration."""
    if ms < 1000:
        return f"{ms}ms"
    s = ms / 1000
    if s < 60:
        return f"{s:.1f}s"
    m = int(s // 60)
    s = s % 60
    return f"{m}m{s:.0f}s"


def generate_markdown(results):
    """Generate a Markdown summary table."""
    lines = [
        "# Public Benchmark Summary",
        "",
        f"Generated: {results[0].get('_timestamp', 'unknown') if results else 'N/A'}",
        "",
        "| Repo | Status | Duration | Tests | Units | FWs | Migration | Portfolio | Determ | Expect |",
        "|------|--------|----------|-------|-------|-----|-----------|-----------|--------|--------|",
    ]

    for r in results:
        status = r["overall"].upper()
        dur = format_duration(r.get("total_duration_ms", 0))
        tests = str(r.get("test_files", "—"))
        units = str(r.get("code_units", "—"))
        fws = str(r.get("frameworks", "—"))
        mig = r.get("migration_readiness", "—")
        port = r.get("portfolio_posture", "—")
        det = r.get("determinism", "—")
        exp = r.get("expectations", "—")
        lines.append(f"| {r['id']} | {status} | {dur} | {tests} | {units} | {fws} | {mig} | {port} | {det} | {exp} |")

    lines.append("")

    # Command-level detail.
    lines.append("## Command Details")
    lines.append("")
    for r in results:
        lines.append(f"### {r['id']}")
        lines.append("")
        lines.append("| Command | Exit | Duration |")
        lines.append("|---------|------|----------|")
        for cmd, info in r.get("commands", {}).items():
            exit_str = "OK" if info["exit_code"] == 0 else f"FAIL({info['exit_code']})"
            lines.append(f"| {cmd} | {exit_str} | {format_duration(info['duration_ms'])} |")
        lines.append("")

        # Posture bands if available.
        if "posture_bands" in r:
            lines.append("**Posture:**")
            for dim, band in r["posture_bands"].items():
                lines.append(f"- {dim}: {band}")
            lines.append("")

        # Migration readiness if available.
        if "migration_readiness" in r:
            lines.append(f"**Migration:** {r['migration_readiness']}"
                         f" ({r.get('migration_blockers', 0)} blockers)")
            if "detected_frameworks" in r:
                lines.append(f"**Frameworks:** {', '.join(r['detected_frameworks'])}")
            lines.append("")

        # Portfolio if available.
        if "portfolio_posture" in r:
            lines.append(f"**Portfolio:** {r['portfolio_posture']}"
                         f" ({r.get('portfolio_assets', '?')} assets)")
            lines.append("")

        # Determinism details if available.
        if "determinism_details" in r:
            lines.append("**Determinism details:**")
            for cmd, status in r["determinism_details"].items():
                lines.append(f"- {cmd}: {status}")
            lines.append("")

    # Warnings.
    warnings = [r for r in results if r["overall"] != "pass"]
    if warnings:
        lines.append("## Warnings")
        lines.append("")
        for r in warnings:
            lines.append(f"- **{r['id']}**: {r['overall']}")
            if r.get("determinism") == "fail":
                lines.append("  - Determinism check failed")
            if r.get("expectations", "").startswith("fail"):
                lines.append("  - Expectation check failed")
            for cmd, info in r.get("commands", {}).items():
                if info["exit_code"] != 0:
                    lines.append(f"  - `{cmd}` exited {info['exit_code']}")
        lines.append("")

    return "\n".join(lines)


def main():
    if not ARTIFACTS.exists():
        print("No artifacts found. Run the benchmark matrix first.", file=sys.stderr)
        sys.exit(1)

    results = []
    from datetime import datetime, timezone

    timestamp = datetime.now(timezone.utc).isoformat()

    for repo_dir in sorted(ARTIFACTS.iterdir()):
        if not repo_dir.is_dir() or repo_dir.name.startswith("."):
            continue
        # Skip non-repo dirs like summary files.
        if not any(repo_dir.glob("*.meta")):
            continue

        entry = summarize_repo(repo_dir)
        entry["_timestamp"] = timestamp
        results.append(entry)

    if not results:
        print("No benchmark results found.", file=sys.stderr)
        sys.exit(1)

    # Print stdout summary.
    print(f"\n{'='*70}")
    print("  Terrain Public Benchmark Summary")
    print(f"{'='*70}\n")

    total_pass = sum(1 for r in results if r["overall"] == "pass")
    total_fail = sum(1 for r in results if r["overall"] == "fail")
    total_degraded = sum(1 for r in results if r["overall"] == "degraded")

    print(f"  Repos: {len(results)}  Pass: {total_pass}  Fail: {total_fail}  Degraded: {total_degraded}\n")

    header = f"  {'Repo':<16} {'Status':<10} {'Duration':<10} {'Tests':<7} {'Units':<7} {'FWs':<5} {'Migrate':<10} {'Portfolio':<10} {'Determ':<8} {'Expect'}"
    print(header)
    print("  " + "-" * (len(header) - 2))

    for r in results:
        status = r["overall"]
        dur = format_duration(r.get("total_duration_ms", 0))
        tests = str(r.get("test_files", "—"))
        units = str(r.get("code_units", "—"))
        fws = str(r.get("frameworks", "—"))
        mig = r.get("migration_readiness", "—")[:9]
        port = r.get("portfolio_posture", "—")[:9]
        det = r.get("determinism", "—")
        exp = r.get("expectations", "—").split("\n")[0][:10]
        print(f"  {r['id']:<16} {status:<10} {dur:<10} {tests:<7} {units:<7} {fws:<5} {mig:<10} {port:<10} {det:<8} {exp}")

    print()

    # Write summary files.
    md = generate_markdown(results)
    md_path = ARTIFACTS / "summary.md"
    md_path.write_text(md)
    print(f"  Written: {md_path}")

    # Clean results for JSON (remove internal keys).
    json_results = []
    for r in results:
        clean = {k: v for k, v in r.items() if not k.startswith("_")}
        json_results.append(clean)

    json_path = ARTIFACTS / "summary.json"
    json_path.write_text(json.dumps(json_results, indent=2) + "\n")
    print(f"  Written: {json_path}")
    print()


if __name__ == "__main__":
    main()
