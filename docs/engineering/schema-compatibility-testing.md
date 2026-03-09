# Schema Compatibility Testing

Schema tests verify that Hamlet's data models survive serialization round-trips
and remain forward-compatible as the schema evolves. Since snapshots are
persisted as JSON (exported files, cached analyses, benchmark data), schema
integrity is critical to the platform's reliability.

## Test Categories

### Round-Trip Testing

Round-trip tests serialize a snapshot to JSON and deserialize it back, then
verify that all fields are preserved. This catches issues like:

- Fields with `json:"-"` that silently drop data
- Pointer fields that deserialize to nil instead of the original value
- Slice fields that deserialize to nil instead of empty slices
- Time fields with timezone or precision loss

```go
func TestSchema_SnapshotRoundTrip(t *testing.T) {
    snap := HealthyBalancedSnapshot()
    data, _ := json.Marshal(snap)
    var decoded models.TestSuiteSnapshot
    json.Unmarshal(data, &decoded)
    // Assert field-by-field equality
}
```

### Forward Compatibility

Forward compatibility tests verify that JSON with unknown fields deserializes
without error. This is essential because:

- Older Hamlet versions may read snapshots produced by newer versions
- External tools may add custom fields to snapshot JSON
- Schema evolution must not break existing consumers

```go
func TestSchema_ForwardCompatibility(t *testing.T) {
    rawJSON := `{
        "repository": {"name": "test-repo"},
        "futureField": "this should be ignored",
        "nested": {"unknown": true}
    }`
    var snap models.TestSuiteSnapshot
    json.Unmarshal([]byte(rawJSON), &snap)  // must not error
}
```

### Versioned Schema Fixtures

Snapshot metadata includes a `SchemaVersion` field. Tests verify that the
version is set correctly on construction and survives serialization. As the
schema evolves, versioned fixtures will allow testing migration paths from
older schema versions.

## Schema Test Inventory

| Test Function | Fixture | What It Validates |
|---|---|---|
| `TestSchema_SnapshotRoundTrip` | HealthyBalancedSnapshot | Full snapshot with frameworks, files, units, ownership |
| `TestSchema_SnapshotWithMeasurements` | MinimalSnapshot + manual measurements | Measurement and posture data round-trip |
| `TestSchema_ForwardCompatibility` | Raw JSON with unknown fields | Unknown field tolerance |
| `TestSchema_EmptySnapshot` | EmptySnapshot | Zero-value and nil-field serialization |

### Planned Schema Tests

| Test Function | Purpose |
|---|---|
| `TestSchema_ImpactAggregateRoundTrip` | Verify impact analysis results survive round-trip |
| `TestSchema_BenchmarkExportRoundTrip` | Verify benchmark export format including posture and quality bands |
| `TestSchema_PortfolioRoundTrip` | Verify portfolio analysis model with investment recommendations |

## Schema Evolution Policy

### Additive Changes (Minor Version)

New fields may be added to existing models within a minor version bump. Rules:

- New fields must have `omitempty` JSON tags so that old JSON without the field
  deserializes cleanly (field gets zero value).
- New fields must not change the semantics of existing fields.
- Forward compatibility tests must continue to pass.
- Golden files must be updated to reflect new fields in output.

### Breaking Changes (Major Version)

Changes that alter or remove existing fields require a major version bump. Rules:

- Renamed fields require a migration path (accept both old and new names).
- Removed fields should be deprecated for at least one minor version first.
- Type changes (e.g., string to int) are always breaking.
- The `SchemaVersion` field must be incremented.

### Practical Guidelines

When adding a new field to `models.TestSuiteSnapshot` or any nested model:

1. Add the field with an `omitempty` JSON tag.
2. Add it to the relevant fixture factory if it should be non-zero in tests.
3. Run `TestSchema_ForwardCompatibility` to confirm old JSON still parses.
4. Run `TestSchema_SnapshotRoundTrip` to confirm the field survives round-trip.
5. Update golden files if the new field appears in any rendered output.

## Relationship to Other Test Categories

- **Golden tests** depend on schema stability -- a schema change that alters JSON
  output will break golden files.
- **Determinism tests** use JSON serialization as their comparison mechanism, so
  they implicitly test that serialization is stable.
- **Adversarial tests** verify behavior with missing or nil fields, which
  complements schema tests that verify structural integrity.

Schema tests are the first line of defense against data loss during persistence
and the foundation for safe schema evolution.
