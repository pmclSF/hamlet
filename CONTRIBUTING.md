# Contributing to Hamlet

## Quick Start

```bash
git clone https://github.com/pmclSF/hamlet.git
cd hamlet
npm install
npm test
```

## Adding a New Framework

### 1. Create Framework Definition

Create `src/languages/{lang}/frameworks/{name}.js`:

```javascript
export default {
  name: 'myframework',
  language: 'javascript',
  paradigm: 'bdd',       // 'bdd', 'xunit', or 'functional'
  detect: {
    imports: [/from ['"]myframework['"]/],
    globals: [/\bmyGlobal\b/],
    patterns: [/myFramework\.specific\(/],
  },
  parse(content) {
    // Return IR (intermediate representation) nodes
    // See existing frameworks for IR node types
  },
  emit(irNodes, options) {
    // Convert IR nodes to target framework code
    // Return { code, imports }
  },
  imports: {
    default: "import { test } from 'myframework';",
  },
};
```

### 2. Register in ConverterFactory

In `src/core/ConverterFactory.js`:

- Add to `FRAMEWORKS` constant
- Add to `FRAMEWORK_LANGUAGE` map
- Add conversion directions to `PIPELINE_DIRECTIONS`

### 3. Write Fixture Tests

Create test fixtures as triplets:

```
test/{lang}/{from}-to-{to}/{category}/{ID}.test.js  # Test file
test/{lang}/{from}-to-{to}/{category}/{ID}.input.ext # Input fixture
test/{lang}/{from}-to-{to}/{category}/{ID}.expected.ext # (optional) Expected output
```

Each test should:
- Import from the direct file path (not barrels)
- Use `ConverterFactory.createConverter(from, to)`
- Test the actual conversion output (no mocks)
- Cover happy path, edge cases, and error conditions

### 4. Add CLI Shorthands

In `src/cli/shorthands.js`:
- Add abbreviation to `FRAMEWORK_ABBREV`
- Add direction entries to `DIRECTIONS`
- Add to appropriate `CONVERSION_CATEGORIES` group

### 5. Add Config Conversion (if applicable)

If the framework has config files, add conversion rules to `src/core/ConfigConverter.js`.

## Code Style

- **ES modules** (`import`/`export`) with `.js` extensions on all relative imports
- **Single quotes**, **2-space indent**, **semicolons always**
- **No mocks** — test real implementations exclusively
- **No `require()`** — except `commitlint.config.js`

## Testing Conventions

- Test file naming: `ClassName.test.js` in matching `test/` subdirectory
- Use `beforeEach` for fresh instances, never share mutable state
- Jest `expect()` assertions only
- Async tests use `async/await`
- Every new public class/function must have test coverage

## Commit Messages

```
type(scope): description

# Types: feat, fix, test, docs, refactor, chore, perf, style, ci, build
```

## Architecture

```
Source Code → Framework Parser → IR Nodes → Framework Emitter → Target Code
                                    ↑
                           ConfidenceScorer
                           (converted/warnings/unconvertible)
```

**IR Node Types:** Suite, Test, Hook, Assertion, Action, Selector, RawCode, Comment

**Key Classes:**
- `ConversionPipeline` — orchestrates parse → convert → emit
- `FrameworkRegistry` — stores framework definitions
- `PipelineConverter` — BaseConverter adapter for the pipeline
- `ConverterFactory` — creates converters (lazy-loads via dynamic import)
- `MigrationEngine` — full project migration with state tracking
