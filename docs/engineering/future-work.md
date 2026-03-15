# Engineering Roadmap — Future Work

This document captures engineering improvements that are designed but not yet
implemented. They are scoped and ready for implementation when the need arises.

## Incremental Analysis (Stage 29)

Today every `terrain analyze` run scans the full repository. For large
monorepos this can be slow. Incremental analysis would skip unchanged files.

**Design:**
- Cache key: file path + content hash (SHA-256 of file bytes)
- Cache location: `.terrain/cache/analysis.json`
- On `--incremental`, load cache, compute hashes for current files, skip
  files whose hash matches
- CLI flags: `--incremental` (use cache), `--no-cache` (force full scan)
- Cache invalidation: any change to `.terrain.yml` policy invalidates all

**What exists today:**
- `models.SortSnapshot()` ensures deterministic output regardless of scan order
- `PipelineDiagnostics` can measure time saved by incremental mode
- File discovery via `filepath.WalkDir` is already the bottleneck

**What would be needed:**
- Content hash computation during `discoverTestFiles`
- Cache serialization and loading
- Merge logic: cached signals for unchanged files + fresh signals for changed
- Tests verifying cache hit/miss behavior

## Parallelization and Profiling (Stage 30)

File content analysis and code unit extraction are embarrassingly parallel
but currently run sequentially. Profiling would identify the actual hotspots.

**Design:**
- Use `sync.WaitGroup` + channel for parallel file reading
- Limit concurrency to `runtime.GOMAXPROCS(0)` workers
- Candidate parallel stages:
  - `analyzeTestFileContent` per file
  - `extractExportedCodeUnits` per file
  - Detector execution (only if detectors are independent)
- Profiling: `go test -bench . -cpuprofile cpu.prof`

**What exists today:**
- `PipelineDiagnostics` provides per-step timing
- All slice accumulation uses append (would need mutex or channel for parallel)
- Detectors run in registration order (some depend on prior signals)

**What would be needed:**
- Benchmark suite for pipeline stages
- Worker pool for file analysis
- Mutex-free result collection via channels
- Tests verifying parallel output matches sequential output

## Layer Separation: Observations vs Derivations (Stage 32)

Today detectors and the risk engine both live in the same pipeline pass.
Separating raw observations (detectors) from derived insights (risk,
recommendations, readiness) would make each layer independently testable.

**Design:**
- **Observation layer**: detectors produce signals from raw snapshot data
- **Derivation layer**: risk engine, readiness model, recommendations consume signals
- Pipeline becomes: analyze → detect → derive → measure
- Each layer has its own test surface

**What exists today:**
- Detectors already implement `signals.Detector` interface
- Risk scoring (`scoring.ComputeRisk`) already runs after detection
- Summary builder (`summary.Build`) already runs after risk scoring
- The separation exists informally; formalizing it would add type safety

**What would be needed:**
- Explicit `ObservationResult` and `DerivationResult` types
- Pipeline split into two phases with typed handoff
- Tests verifying observation layer produces no derived data
- Tests verifying derivation layer consumes only signals, not raw files
