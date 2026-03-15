# Test Identity Model

Terrain assigns every discovered test case a **deterministic, stable, persistent test ID** suitable for snapshots, comparisons, trend analysis, flake tracking, and coverage attribution.

## Why Stable Identity Matters

Without stable identity, Terrain cannot:
- Track tests across snapshots (trend analysis)
- Detect flaky tests over time
- Attribute coverage to specific tests
- Detect test additions, removals, or renames between runs

## Architecture

```
internal/identity/       Normalization and hashing primitives
internal/testcase/       Test case model, extraction, collision detection
internal/models/         TestCase model for snapshot serialization
```

## Canonical Identity

Each test's identity is constructed from structural properties:

```
{normalized_path}::{suite_hierarchy}::{test_name}[::param_signature]
```

Example:
```
src/__tests__/auth.test.js::AuthService > login::should return a token
```

The `identity.BuildCanonical()` function constructs this from its components. The `identity.ParseCanonical()` function reverses it.

### What IS part of identity

- **File path** (repository-relative, slash-normalized)
- **Suite hierarchy** (describe/class nesting chain)
- **Test name** (the `it`/`test`/function name)
- **Parameter signature** (when statically available and stable)

### What is NOT part of identity

- **Line numbers** — stored as metadata only, never hashed
- **Random UUIDs** — forbidden
- **Traversal/discovery order** — forbidden
- **Runtime execution order** — forbidden

## Path Normalization

Rules applied by `identity.NormalizePath()`:

| Rule | Example |
|------|---------|
| Convert backslashes to forward slashes | `src\test\a.js` → `src/test/a.js` |
| Strip leading `./` | `./src/test.js` → `src/test.js` |
| Strip leading `/` | `/src/test.js` → `src/test.js` |
| Preserve case | Case-sensitive (Linux semantics) |
| Strip invalid UTF-8 | Invalid byte sequences removed |

## Name Normalization

Rules applied by `identity.NormalizeName()`:

| Rule | Example |
|------|---------|
| Trim leading/trailing whitespace | `"  test  "` → `"test"` |
| Collapse internal whitespace to single space | `"should  do\tsomething"` → `"should do something"` |
| Strip invalid UTF-8 | Invalid byte sequences removed |
| Preserve case | Case-sensitive |

Suite hierarchies are normalized per-element, then joined with ` > `.

## Test ID Generation

`identity.GenerateID()` produces a deterministic 16-hex-character ID from a canonical identity string using truncated SHA-256:

```go
func GenerateID(canonical string) string {
    h := sha256.Sum256([]byte(canonical))
    return hex.EncodeToString(h[:])[:16]
}
```

Properties:
- **Deterministic**: same canonical identity → same ID, always
- **Stable**: independent of traversal order or runtime state
- **Compact**: 16 hex characters (64 bits of entropy)

## Extraction

`testcase.Extract()` discovers test cases from source files by language:

| Language | Patterns |
|----------|----------|
| JS/TS | `describe`/`it`/`test` with brace-counting scope tracking, `test.each`/`it.each` for parameterized |
| Go | `func Test*` top-level, `t.Run()` subtests |
| Python | `class Test*`, `def test_*`, `@pytest.mark.parametrize` |
| Java | `@Test`/`@ParameterizedTest` annotations, class scope |

Each extracted test case gets:
- `ExtractionKind`: `static`, `parameterized_template`, `dynamic`, `ambiguous`
- `Confidence`: 0.0–1.0 reflecting extraction quality

## Collision Detection

`testcase.DetectAndResolveCollisions()` handles the case where two tests produce identical canonical identities (e.g., duplicate test names in the same suite). Resolution:

1. Group by canonical identity
2. Sort collisions by line number (deterministic)
3. First occurrence keeps original identity
4. Subsequent occurrences get `#N` suffix appended

Collision diagnostics are emitted for visibility.

## Parameterized Tests

Policy: prefer **template-level identity** over instance-level identity.

- `test.each(...)('name', ...)` → extracted as `parameterized_template` with the template name
- `@pytest.mark.parametrize` → extracted as `parameterized_template`
- `@ParameterizedTest` → extracted as `parameterized_template`

Parameter signatures are included in identity only when statically available and stable. Dynamic tests that cannot be reliably identified are marked `ExtractionKind: "dynamic"` or `"ambiguous"` with reduced confidence.

## Integration

Test case extraction and identity assignment happen during static analysis in the pipeline:

```
engine.RunPipeline → analysis.Analyze() → testcase.Extract() per file
                                         → testcase.DetectAndResolveCollisions()
                                         → testtype.InferAll()
                                         → snapshot.TestCases populated
```

## Test Guarantees

The test suite verifies:
- Same source → same ID (deterministic)
- Reordered extraction → same IDs (order-independent)
- Whitespace normalization stability
- Path normalization stability
- Line movement without rename → same ID
- Rename of test name or suite → new ID
- Collision detection and deterministic resolution
