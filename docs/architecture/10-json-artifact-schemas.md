# JSON Artifact Schemas

> **Status:** Implemented
> **Purpose:** Define the standard JSON envelope and per-engine result schemas for all Terrain artifacts
> **Key decisions:**
> - All artifacts share a common envelope with version, repo, SHAs, and timestamp
> - Schema versioning is per-artifact; additive field additions are non-breaking
> - Artifacts are written to a fixed path (`.terrain/artifacts/`) with deterministic filenames per engine
> - Envelope stability is guaranteed; result payloads may evolve with version bumps

See also: [07-pr-ci-integration.md](07-pr-ci-integration.md), [09-cli-spec.md](09-cli-spec.md)

## Artifact Envelope

All JSON artifacts share a common envelope:

```json
{
  "version": "1.0.0",
  "repo": "repository-name",
  "base_sha": "abc123def456",
  "head_sha": "789012fed345",
  "generated_at": "2025-01-15T10:30:00.000Z",
  "results": { }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Artifact schema version |
| `repo` | string | Repository name (from git remote or directory name) |
| `base_sha` | string | Base commit SHA |
| `head_sha` | string | Head commit SHA |
| `generated_at` | string | ISO 8601 timestamp |
| `results` | object | Engine-specific payload |

## Impact Artifact

File: `.terrain/artifacts/terrain-impact.json`

```json
{
  "version": "1.0.0",
  "repo": "my-app",
  "base_sha": "abc123",
  "head_sha": "def456",
  "generated_at": "2025-01-15T10:30:00.000Z",
  "results": {
    "changedFiles": ["src/auth/login.ts", "src/utils/validate.ts"],
    "impactedTests": [
      {
        "testId": "test:tests/auth/login.test.ts:10:should validate credentials",
        "confidence": 0.85,
        "impactChain": [
          "file:src/auth/login.ts",
          "file:tests/auth/login.test.ts",
          "test:tests/auth/login.test.ts:10:should validate credentials"
        ]
      }
    ],
    "testCount": 5
  }
}
```

## Coverage Artifact

File: `.terrain/artifacts/terrain-coverage.json`

```json
{
  "version": "1.0.0",
  "results": {
    "sourceCount": 42,
    "bandCounts": { "High": 20, "Medium": 12, "Low": 10 },
    "sources": [
      {
        "sourceId": "file:src/auth/login.ts",
        "path": "src/auth/login.ts",
        "testCount": 5,
        "directTests": ["test:tests/auth/login.test.ts:10:should validate"],
        "indirectTests": ["test:tests/integration/auth.test.ts:20:full flow"],
        "band": "High"
      }
    ]
  }
}
```

## Duplicates Artifact

File: `.terrain/artifacts/terrain-duplicates.json`

```json
{
  "version": "1.0.0",
  "results": {
    "testsAnalyzed": 150,
    "duplicateCount": 12,
    "clusters": [
      {
        "members": [
          "test:tests/auth/login.test.ts:10:should validate",
          "test:tests/auth/login-alt.test.ts:8:should validate"
        ],
        "scores": {
          "fixtureOverlap": 1.0,
          "helperOverlap": 0.8,
          "suitePathSimilarity": 0.9,
          "assertionPatternSimilarity": 0.7
        },
        "overallScore": 0.86
      }
    ]
  }
}
```

## Stability Guarantees

- The envelope format (`version`, `repo`, `base_sha`, `head_sha`, `generated_at`) is stable
- The `version` field will be incremented if breaking changes are made to a result schema
- New fields may be added to result objects without a version bump (additive changes are non-breaking)
- Field removal or type changes require a version bump

## Artifact Storage

Artifacts are written to `.terrain/artifacts/` by default. The directory is created automatically. Each engine writes to a fixed filename, overwriting previous artifacts.

```
.terrain/
  artifacts/
    terrain-impact.json
    terrain-coverage.json
    terrain-duplicates.json
```
