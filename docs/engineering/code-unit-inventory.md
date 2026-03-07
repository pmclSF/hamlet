# Code Unit Inventory

Hamlet maintains a normalized inventory of code units — functions, methods, classes, and modules — so coverage can be attributed to meaningful code structure rather than raw lines.

## CodeUnit Model

```go
type CodeUnit struct {
    UnitID         string       // deterministic stable ID: "path:symbol" or "path:parent.symbol"
    Name           string       // local identifier
    Path           string       // repository-relative file path
    Kind           CodeUnitKind // function, method, class, module, unknown
    Exported       bool         // externally visible
    ParentName     string       // containing class/struct for methods
    Language       string       // programming language
    StartLine      int          // definition start line (metadata)
    EndLine        int          // definition end line (metadata)
    Complexity     float64      // optional complexity estimate
    Coverage       float64      // optional coverage ratio
    LinkedTestFiles []string    // associated test files
    Owner          string       // resolved owner
}
```

### Unit ID Construction

Unit IDs follow the format `normalized_path:symbol_name` (or `path:parent.symbol` for methods):

```
src/utils.js:formatDate
src/api.js:ApiClient.fetchData
internal/scoring/risk_engine.go:ComputeRisk
```

IDs are deterministic and stable across runs for the same source structure.

### CodeUnit Kinds

| Kind | Description |
|------|-------------|
| `function` | Top-level function or standalone function |
| `method` | Method on a class/struct |
| `class` | Class or struct definition |
| `module` | Module-level export |
| `unknown` | Could not classify |

## Extraction

Code unit extraction is integrated into `analysis.Analyze()` via the `LanguageAnalyzer` interface:

```go
type LanguageAnalyzer interface {
    Language() string
    CountTests(src string) int
    CountAssertions(src string) int
    CountMocks(src string) int
    CountSnapshots(src string) int
    ExtractExports(root, relPath string) []models.CodeUnit
}
```

### Language Support

| Language | Functions | Methods | Classes | Exports |
|----------|-----------|---------|---------|---------|
| JS/TS | ✅ | ✅ | ✅ | ✅ (export keyword) |
| Go | ✅ | ✅ (receiver methods) | ✅ (structs) | ✅ (capitalized) |
| Python | ✅ | ✅ | ✅ | ✅ (no underscore prefix) |
| Java | ✅ | ✅ | ✅ | ✅ (public keyword) |

### Extraction Rules

- **JS/TS**: Exported functions, class methods, module-level arrow functions
- **Go**: Exported (capitalized) functions, struct methods with receivers
- **Python**: Top-level functions, class methods, public symbols (no `_` prefix)
- **Java**: Public classes, public/protected methods

## Usage

Code units serve as the join point between:
- **Coverage records** (line/function hits map to code unit spans)
- **Test cases** (test IDs can be linked to covered unit IDs)
- **Signals** (untested-export detector checks for units with no linked tests)
- **Risk surfaces** (coverage gaps contribute to risk scoring)

## Limitations

- Extraction is regex-based, not AST-based (pragmatic accuracy over perfection)
- Start/end line estimates may be approximate for complex nested structures
- Dynamic exports or computed property names are not captured
