# Orphaned Code Provenance Pass (First Guess)

Date: 2026-03-10
Scope: Go CLI (`cmd/terrain`), benchmark runners, TypeScript graph engine surface, npm package/release wiring.

This pass is a triage map, not a delete list. Every candidate below has:
- evidence of current usage,
- likely canonical owner,
- first-guess action (`keep`, `move`, `merge`, `delete`, `decision-required`),
- safety checks before any removal.

---

## Decision Matrix

| Candidate | Current Evidence | Likely Canonical Owner | First Guess | Confidence | Safety Checks Before Change |
|---|---|---|---|---|---|
| `benchmarks/cli/*` (standalone benchmark runner) | Invoked only by `benchmarks/run-cli-benchmarks.sh`; docs call it "legacy wrapper". Near-duplicate logic exists in `internal/benchmark/*` + `cmd/terrain-bench`. | `internal/benchmark/*` + `cmd/terrain-bench` | `merge` then `delete` legacy runner | High | Diff parity of outputs on same repo set; add compatibility wrapper in script; update docs and CI references. |
| `benchmarks/run-cli-benchmarks.sh` | Still used as shell entrypoint; currently executes `go run ./benchmarks/cli/`. | `cmd/terrain-bench` | `keep` as compatibility shim, repoint target | High | Repoint to `go run ./cmd/terrain-bench/`; verify flags parity with one smoke run. |
| `internal/analysis/content_analysis.go::analyzeTestFileContent` | Unused wrapper; only `analyzeTestFileContentCached` is called by analyzer. | `internal/analysis/analyzer.go` path (cached variant) | `delete` wrapper | High | Run `go test ./internal/analysis ./internal/engine`; re-run `staticcheck`. |
| `internal/engine/pipeline.go::ingestRuntime` | Unused wrapper; context-aware ingestion path is used. | `RunPipelineContext` helpers (`ingestRuntimeArtifacts` + `applyRuntimeResults`) | `delete` wrapper | High | Run `go test ./internal/engine ./internal/runtime ./internal/health`; verify no symbol reference remains. |
| `internal/engine/pipeline.go::ingestCoverage` | Unused wrapper; context-aware ingestion path is used. | `RunPipelineContext` helpers (`ingestCoverageArtifacts` + `applyCoverageArtifacts`) | `delete` wrapper | High | Run `go test ./internal/engine ./internal/coverage`; verify coverage ingest behavior unchanged. |
| Benchmark explain strategy drift (`benchmarks/cli/runner.go` vs `internal/benchmark/command_executor.go`) | `benchmarks/cli` uses impact→show path; `internal/benchmark` uses analyze→impact→explain path. Same command name, different semantics. | `internal/benchmark` (shared package used by `cmd/terrain-bench`) | `merge` behavior into one implementation; retire duplicate | High | Golden benchmark output comparison before/after on 2 repos; update docs for exact explain strategy. |
| `cmd/terrain-bench` log text: "parallel per repo" | Runner is now sequential within repo, but printed text still says parallel per repo. | `cmd/terrain-bench` | `keep`, fix stale text | High | One benchmark run confirms output messaging. |
| `cmd/terrain/main.go` stale signal branch `conditionallySkippedTest` | Signal catalog does not define this type; skipped detector emits `skippedTest`. | Signal catalog + health detector set | `delete` stale branch or formalize signal type | High | If deleting: run `go test ./cmd/terrain ./internal/health ./internal/models`. |
| npm bin target (`package.json` `"terrain": "dist/cli/index.js"`) vs package contents (`files`) | Packed artifact excludes required `dist/graph`, `dist/analysis`, etc.; installed CLI fails `ERR_MODULE_NOT_FOUND`. | Depends on product decision (converter CLI vs graph CLI) | `decision-required` | High | Choose one: (A) converter bin (`bin/terrain.js`) OR (B) include all dist graph modules in package files and verify pack. |
| TS graph engine source (`src/analysis`, `src/graph`, `src/engines`, `src/parsers`, `src/artifacts`, `src/cli/index.ts`) | Fully implemented and tested, but not coherently shipped via npm today. Docs and package semantics are mixed with converter surface. | Separate product lane or explicit subcommand package contract | `decision-required` (not orphan, but misplaced ownership) | Medium | Define ownership contract: "experimental internal", "published as default CLI", or "separate package". Then align `bin`, `files`, docs, and release gates. |
| `docs/cli-spec.md` TS graph command section under primary CLI spec | Documents `terrain graph ...` requiring TS build; primary Go CLI usage/docs point elsewhere. | Docs split by product surface | `move` section to explicit TS/experimental docs or separate CLI spec | Medium | Validate doc links from README; ensure no broken command examples for default install path. |
| `.eslintrc.json` without TS parser while linting `src/**/*.ts` | `npm run lint` fails on every TS file parse. | Node toolchain config | `keep`, but reconfigure | High | Add TS parser/config; run `npm run lint` and `npm run release:verify`. |

---

## Non-Orphan But High-Risk "Looks Orphaned" Items

These are active, but currently look abandoned due to product wiring issues.

- TypeScript graph CLI runtime (`dist/cli/index.js`) is buildable and runnable locally, but effectively orphaned in the published package because required runtime modules are excluded from `npm` `files`.
- `cmd/terrain` debug/insight heavy paths are active but operationally fragile (very high runtime on large repos). This is performance debt, not dead code.

---

## Execution Order (Safe Cleanup)

1. Consolidate benchmark runners first (`benchmarks/cli` -> `internal/benchmark` + `cmd/terrain-bench`).
2. Remove confirmed dead wrappers/functions (`analyzeTestFileContent`, `ingestRuntime`, `ingestCoverage`).
3. Resolve npm/bin ownership decision (converter CLI vs TS graph CLI).
4. Align docs and release gates to the chosen ownership model.
5. Run full verification:
   - `go test ./...`
   - `go vet ./...`
   - `staticcheck ./...`
   - `npm run lint`
   - `npm test`
   - `npm run release:verify`

---

## Notes

- This matrix intentionally distinguishes "unused and deletable now" from "misowned and needs product decision."
- Any delete should happen only after the corresponding safety check passes.
