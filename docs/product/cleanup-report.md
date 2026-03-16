# Deprecated Functionality Cleanup Report

> **Date:** 2026-03-15
> **Scope:** Remove obsolete CLI commands, unused graph types, orphaned packages, and Hamlet-era naming.

## Summary

| Category | Items Removed | Lines Deleted |
|----------|--------------|---------------|
| Orphaned Go packages | 5 packages | ~1,800 |
| Hamlet-era aliases | 3 items (bin/hamlet.js, Go detection, package.json bin) | ~15 |
| Stale graph schema doc | 1 doc replaced | ~275 (replaced with redirect) |
| Benchmark artifact cleanup | ~70 files updated | (sed replacements) |

**Total dead code removed:** ~1,815 lines across 5 packages + 1 file.

---

## 1. Orphaned Go Packages Removed

Five packages in `internal/` were fully implemented with tests but **never imported by any other package**. They were prototyped features that were either superseded by other implementations or never integrated.

| Package | Purpose | Files | Estimated Lines | Superseded By |
|---------|---------|-------|-----------------|---------------|
| `internal/assertion/` | Assertion quality assessment | 3 | ~400 | Quality signal detectors in `internal/quality/` |
| `internal/clustering/` | Test clustering detection | 3 | ~300 | Stability clustering in `internal/stability/` |
| `internal/envdepth/` | Environment depth assessment | 3 | ~350 | Environment matrix in `internal/matrix/` |
| `internal/failure/` | Failure classification taxonomy | 3 | ~400 | Health detectors in `internal/health/` |
| `internal/suppression/` | Signal suppression rules | 3 | ~350 | Not replaced (feature deferred) |

**Verification:** `grep -r "terrain/internal/<pkg>" --include="*.go"` returned zero matches for each package before deletion.

**Previously removed (Prompt 35):** `internal/reasoning/` (~1,300 lines) — superseded by domain-specific reasoning in depgraph, impact, stability, and matrix packages.

---

## 2. Hamlet-Era Naming Removed

The product was renamed from Hamlet to Terrain on 2026-03-13. Deprecation aliases were maintained for backward compatibility. This cleanup removes all remaining aliases.

### Removed

| Item | Location | What It Did |
|------|----------|-------------|
| `bin/hamlet.js` | Root | Deprecated JS CLI alias — printed warning and re-exported `bin/terrain.js` |
| Hamlet binary detection | `cmd/terrain/main.go:83-88` | Detected invocation as `hamlet`/`hamlet.exe` and printed deprecation warning |
| `"hamlet"` npm bin entry | `package.json:28` | Mapped `hamlet` command to `bin/hamlet.js` |

### Remaining (filesystem only)

The repository checkout directory is still named `hamlet/` on the local filesystem. This is the user's clone path, not a product name reference. Filesystem paths in benchmark artifacts (`/Users/pzachary/hamlet/...`) reflect this and are not stale naming.

---

## 3. Stale Graph Schema Documentation

### `docs/architecture/02-graph-schema.md` — Replaced with redirect

The original 289-line document described 26 node types and 22 edge types. After graph schema hardening (Prompt 33), 6 node types and 7 edge types were removed from the implementation but the doc was not updated.

**Replaced with:** A redirect to the authoritative `16-unified-graph-schema.md`, which documents the current 20 node types and 15 edge types and lists the removed types for reference.

### Types that were removed from implementation (Prompt 33)

**Node types (6):** `package`, `service`, `generated_artifact`, `external_service`, `fixture`, `helper`

**Edge types (7):** `test_uses_fixture`, `test_uses_helper`, `fixture_imports_source`, `helper_imports_source`, `validates`, `test_exercises`, `depends_on_service`

---

## 4. Benchmark Artifact Cleanup

~70 benchmark output files in `artifacts/public-benchmarks/` contained stale references:

| Pattern | Occurrences | Replacement |
|---------|-------------|-------------|
| `.hamlet/policy.yaml` | ~70 files | `.terrain/policy.yaml` |
| `hamlet analyze` | ~40 files | `terrain analyze` |
| `hamlet posture`, `hamlet summary`, etc. | ~30 files | `terrain posture`, `terrain summary`, etc. |

Remaining `hamlet` strings in artifacts are filesystem paths (`/Users/pzachary/hamlet/...`) which reflect the local checkout directory, not product naming.

---

## 5. Obsolete CLI Commands — None Found

All CLI commands in `cmd/terrain/main.go` have valid handler functions and produce real output. The `terrain depgraph` backward-compatibility alias for `terrain debug depgraph` remains intentional and documented.

**Commands verified (27 total):**
- 4 canonical: analyze, impact, insights, explain
- 13 supporting: init, summary, focus, posture, portfolio, metrics, compare, migration, policy check, select-tests, pr, show, export benchmark
- 5 AI: ai list, ai doctor, ai run (scaffolded), ai record (scaffolded), ai baseline (scaffolded)
- 5 debug: debug graph, debug coverage, debug fanout, debug duplicates, debug depgraph
- Plus: version, help

The 3 scaffolded AI commands (`ai run`, `ai record`, `ai baseline`) return clear not-implemented messages with documentation pointers. These are intentional placeholders, not obsolete commands.

---

## 6. Unused Graph Types — None Remaining

After Prompt 33's graph schema hardening:
- `grep -r "NodePackage\|NodeService\|NodeGeneratedArtifact\|NodeExternalService\|NodeFixture\|NodeHelper" --include="*.go" internal/` — **zero matches**
- `grep -r "EdgeTestUsesFixture\|EdgeTestUsesHelper\|EdgeFixtureImportsSource\|EdgeHelperImportsSource\|EdgeValidates\|EdgeTestExercises\|EdgeDependsOnService" --include="*.go" internal/` — **zero matches**

All removed types are absent from the codebase.

---

## What Remains (Intentional)

| Item | Reason |
|------|--------|
| `terrain depgraph` alias | Backward compatibility for `terrain debug depgraph`. Documented. |
| AI scaffolded commands | `ai run`, `ai record`, `ai baseline` — planned features with clear not-implemented messages |
| Graph reserved types | `NodeValidationTarget`, `NodeExecutionRun`, `NodeValidationExecution`, `NodePrompt`, `NodeDataset`, `NodeModel`, `NodeEvalMetric` — defined in schema for future use |

---

## Verification

- Go build: clean (0 errors)
- Go tests: 36 packages pass (down from 41 — 5 deleted packages)
- Golden snapshot tests: all 4 pass
- No import errors from deleted packages
