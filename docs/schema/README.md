# Terrain JSON Schema Contracts

This directory contains JSON Schema definitions for Terrain's machine-readable outputs.

## Schemas

| Schema | Produced by | Purpose |
|--------|------------|---------|
| `analysis.schema.json` | `terrain analyze --json` | Framework detection and project scan results |
| `conversion.schema.json` | `terrain convert --report-json <file>` | Structured conversion run report |

## Versioning

Each schema includes a `schemaVersion` field (e.g., `"1.0.0"`).

- **Patch** (1.0.x): Documentation-only changes, no field changes.
- **Minor** (1.x.0): New optional fields added. Existing consumers are unaffected.
- **Major** (x.0.0): Required fields added/removed, field types changed, or fields renamed. Consumers must update.

## Backward Compatibility

- Required fields listed in the schema will not be removed without a major version bump.
- New optional fields may be added in minor releases.
- Field types will not change without a major version bump.
- Enum values may be extended (new values added) in minor releases.

## Consuming Reports

```bash
# Analysis: pipe JSON to stdout
terrain analyze src/ --json | jq '.summary.frameworksDetected'

# Analysis: write to file
terrain analyze src/ --out analysis.json

# Conversion: write report alongside converted files
terrain convert tests/ --from jest --to vitest -o out/ --report-json report.json
```
