# Hosted / Organization Product Boundaries

This document identifies what remains future work for a hosted, multi-org
product. It exists to prevent accidentally claiming these features are
shipped and to guide future architecture decisions.

## What Ships Today (Local-Only)

Everything Hamlet does today runs on a single machine against a single repo:

- Static analysis and signal detection
- Runtime and coverage artifact ingestion
- Explainable risk modeling
- Snapshot persistence and local trend comparison
- Migration readiness and preview
- Local policy and governance
- Benchmark-safe export (no hosted comparison)
- Executive summaries
- VS Code extension (reads local CLI output)

## What Does NOT Ship Today

### Cross-Repo Aggregation

The metrics model and benchmark export schema are designed to support
aggregation, but no aggregation service exists.

**What would be needed:**
- Ingest service accepting benchmark exports from multiple repos
- Aggregation engine producing org-level summaries
- Privacy guarantees (no raw paths in transit)
- Storage model for historical aggregates

### Hosted Benchmark Comparison

`hamlet export benchmark` produces a privacy-safe artifact. Today it is
a local file. Future comparison requires:

**What would be needed:**
- Benchmark database with segmentation (size, language, framework mix)
- Percentile computation across segments
- API for submitting and querying benchmarks
- Anonymization guarantees

### Auth / Accounts / Teams

Hamlet has no identity layer.

**What would be needed:**
- Authentication (OAuth, SSO)
- Organization and team model
- Role-based access to aggregated data
- API key management for CI integrations

### Org Dashboards

No web UI exists.

**What would be needed:**
- Dashboard frontend consuming aggregated data
- Portfolio views (risk posture across repos)
- Trend visualizations (org-level)
- Team-level views (rollups by CODEOWNERS)

### Portfolio Views

The ownership model supports per-team attribution. Portfolio views
would aggregate this across repos.

**What would be needed:**
- Cross-repo ownership normalization
- Team health summaries
- Migration progress tracking across org

### CI Annotation Service

`hamlet policy check` returns exit codes for CI. A hosted service could
provide richer PR annotations.

**What would be needed:**
- GitHub/GitLab integration for PR comments
- Annotation format (inline findings, summary comments)
- Diff-aware filtering (only annotate changed lines)

## Architecture Principles for Future Work

1. **Snapshot-first**: All future aggregation should consume `TestSuiteSnapshot` or `BenchmarkExport` artifacts. Do not bypass the snapshot boundary.

2. **Privacy boundary holds**: Aggregate services must never store raw file paths or source code. The benchmark export schema enforces this.

3. **Local value must not degrade**: Hosted features are additive. Hamlet must remain fully useful without network access, accounts, or SaaS.

4. **Extension stays thin**: If a dashboard exists, the extension should link to it — not reimplement it.
