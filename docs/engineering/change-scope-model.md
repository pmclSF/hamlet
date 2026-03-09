# Change Scope Model

Stage 122 -- How Hamlet determines what changed in the repository.

## Overview

`ChangeScope` is the input to impact analysis. It normalizes changed-file information from multiple sources into a uniform model that `Analyze()` consumes. All adapters produce the same `ChangeScope` struct regardless of origin.

Defined in `internal/impact/impact.go` and constructed by adapters in `internal/impact/changescope.go`.

## ChangeScope Model

```go
type ChangeScope struct {
    ChangedFiles []ChangedFile  // Files that were added, modified, deleted, or renamed
    BaselineRef  string         // Git ref or snapshot used as baseline (optional)
    CurrentRef   string         // Git ref representing current state (optional)
    Source       string         // How the scope was determined
}
```

`Source` is one of: `"git-diff"`, `"explicit"`, `"ci-changed-files"`, `"snapshot-compare"`.

## ChangedFile

```go
type ChangedFile struct {
    Path       string      // Repo-relative file path
    ChangeKind ChangeKind  // added, modified, deleted, renamed
    OldPath    string      // Set only when ChangeKind is "renamed"
    IsTestFile bool        // True if the file is detected as a test file
}
```

## Input Adapters

### ChangeScopeFromGitDiff

Runs `git diff --name-status <baseRef>` against the repository root. Parses git status codes (`A`, `M`, `D`, `R`) into `ChangeKind` values. Defaults to `HEAD~1` when no base ref is provided.

```go
scope, err := ChangeScopeFromGitDiff("/path/to/repo", "origin/main")
```

### ChangeScopeFromPaths

Creates a scope from an explicit list of file paths with a uniform change kind. Used for programmatic or API-driven invocations.

```go
scope := ChangeScopeFromPaths([]string{"src/auth.js", "src/db.js"}, ChangeModified)
```

### ChangeScopeFromCIList

Parses a newline-separated list of changed file paths, as typically provided by CI environment variables (e.g., `CHANGED_FILES`). All files are treated as modified since CI lists usually do not include status information.

```go
scope := ChangeScopeFromCIList(os.Getenv("CHANGED_FILES"), repoRoot)
```

### ChangeScopeFromComparison

Compares two `TestSuiteSnapshot` values to identify files that were added, removed, or changed between snapshots. Files present only in the target snapshot are marked added; files present only in the baseline are marked deleted; files present in both are assumed modified (content-level diffing is not performed).

```go
scope := ChangeScopeFromComparison(baselineSnap, currentSnap)
```

Results are sorted by path for deterministic output.

## Path Normalization

All adapters normalize paths to repo-relative, forward-slash form:

- Absolute paths are converted to relative using `filepath.Rel(repoRoot, path)`.
- `ChangeScopeFromCIList` additionally applies `filepath.ToSlash(filepath.Clean(path))` to handle OS-specific separators and redundant components.
- Renamed files store the original path in `OldPath` and the new path in `Path`.

## Change Kinds

| Kind | Value | Meaning |
|------|-------|---------|
| Added | `"added"` | New file introduced |
| Modified | `"modified"` | Existing file changed |
| Deleted | `"deleted"` | File removed |
| Renamed | `"renamed"` | File moved or renamed; `OldPath` is set |

## Test File Detection Heuristics

`isTestFilePath()` classifies a path as a test file using two strategies:

**Filename patterns** -- the lowercased base name is checked for: `.test.`, `.spec.`, `_test.`, `_spec.`, `test_`, `spec_`. Go test files (`_test.go` suffix) are matched explicitly.

**Directory patterns** -- the full lowercased path is checked for known test directories: `/test/`, `/tests/`, `/__tests__/`, `/e2e/`, `/cypress/`, `/spec/`.

A file matching any pattern is marked `IsTestFile: true` and excluded from code-unit mapping (test files are handled separately during impacted-test discovery).

## Confidence of Change Mapping

The `ChangeScope` itself does not carry confidence -- it is a factual record of what changed. Confidence is assigned downstream when `mapChangedUnits()` maps changed files to code units:

- **Exact** -- file was added or deleted, so all units in it are definitively affected.
- **Inferred** -- file was modified, so units are assumed affected (without line-level diffing).
- **Weak** -- no code units were parsed for the file; a file-level fallback unit is created.
