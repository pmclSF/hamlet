# Hamlet v2 — Multi-Framework Test Converter Architecture

## 1. Architecture Decision Record

### Problem

Hamlet currently converts between 3 JavaScript E2E frameworks (Cypress, Playwright, Selenium) using 6 dedicated converter classes. Each converter is a standalone 300–640 line class that mixes structural transformation with API-level pattern matching. Expanding to 30 conversion directions across 5 languages would require ~30 such classes under the current approach — roughly 12,000–18,000 lines of duplicated conversion logic.

### Options Evaluated

**Option A: Intermediate Representation (IR)**

Each framework gets a Parser (source → IR) and an Emitter (IR → target). The IR is a normalized test AST capturing suites, tests, hooks, assertions, and actions.

- 30 converters become ~20 parsers + ~20 emitters (linear growth, not quadratic)
- Handles structural transforms cleanly (pytest functions → unittest classes)
- The IR must be expressive enough for all paradigms — BDD, xUnit, function-based
- Lossy for framework-specific features (Cypress custom commands, RSpec `let!`)
- Entire test file must be fully parsed before any output, which makes partial conversion harder

**Option B: Grouped Pattern Registry**

Frameworks grouped by structural similarity (jest-like, xunit-like, pytest-like). Within a group: simple pattern mapping. Across groups: structural templates.

- Simple for near-identical frameworks (Jest ↔ Vitest is nearly 1:1)
- Cross-group transforms (pytest → unittest) require structural templates that are effectively a second system
- "Group" boundaries are fuzzy — Mocha with Chai is structurally like Jest but the assertion library is completely different
- Doesn't solve the fundamental scaling problem for cross-paradigm conversions

**Option C: Hybrid (IR + PatternEngine)**

IR for test structure (suites, hooks, lifecycle, imports). PatternEngine for API-level transforms (assertions, mocking, selectors).

- Structural transforms (wrapping pytest functions in unittest classes, converting RSpec `let` to Minitest instance variables) go through the IR
- API-level transforms (`.should('be.visible')` → `.toBeVisible()`, `assert x == y` → `self.assertEqual(x, y)`) go through the PatternEngine
- The existing codebase already separates these concerns informally — `CypressToPlaywright.convert()` calls `convertCypressCommands()` (API-level) then `convertTestStructure()` (structural) then `transformTestCallbacks()` (structural) as three distinct phases
- Two systems to maintain, but each system is simpler than a single monolithic approach

### Recommendation: Option C (Hybrid)

The existing codebase already implements the hybrid approach without formalizing it. Evidence from `src/converters/CypressToPlaywright.js`:

```
convert(content) {
  result = this.convertCypressCommands(result);    // API-level (PatternEngine territory)
  result = this.convertTestStructure(result);       // Structural (IR territory)
  result = this.transformTestCallbacks(result);     // Structural (IR territory)
  const imports = this.getImports(testTypes);       // Structural (IR territory)
  result = this.cleanupOutput(result);              // Post-processing
}
```

The hybrid formalizes this separation:

1. **IR handles structure**: Test file skeleton — suites, tests, hooks, imports, class wrapping, indentation. This is where paradigm shifts happen (BDD → xUnit, function-based → class-based).

2. **PatternEngine handles API transforms**: Assertion syntax, command mapping, selector conversion, mocking API translation. These are regex-replaceable because they operate on individual expressions, not on document structure.

3. **Structural transforms that regex can't handle** (the hard cases):
   - pytest: `def test_foo():` with module-level fixtures → unittest: `class TestFoo(unittest.TestCase):` with `setUp`/`tearDown` methods
   - RSpec: `let(:user) { create(:user) }` inside `describe` → Minitest: `def setup; @user = create(:user); end` inside `class`
   - JUnit 4: `@Test(expected = IOException.class)` → JUnit 5: `assertThrows(IOException.class, () -> { ... })` wrapping the method body in a lambda

   These require understanding the nesting structure of the test file, which regex alone cannot reliably do.

4. **Why full AST parsing is overkill**: We don't need to understand every language construct — only test-related structures (describe/it/class/def, hooks, assertions, imports). A lightweight structural parser that identifies these constructs and passes everything else through as raw code is sufficient. Tree-sitter provides this capability for all target languages from Node.js without requiring language runtimes.

---

## 2. IR Schema

The IR represents a normalized test file as a tree of typed nodes. It is intentionally minimal — it captures test structure and metadata, not general-purpose code semantics. Code that isn't test structure passes through as `RawCode` nodes.

### Node Types

```
TestFile
├── language: 'javascript' | 'python' | 'ruby' | 'java'
├── imports: ImportStatement[]
├── body: (TestSuite | TestCase | RawCode)[]
└── sourceMap: Map<nodeId, { startLine, endLine }>

TestSuite
├── name: string
├── hooks: Hook[]
├── tests: (TestCase | TestSuite)[]      // supports nesting
├── sharedState: SharedVariable[]         // let/subject/instance vars
└── modifiers: Modifier[]                 // skip, only, tags

TestCase
├── name: string
├── body: (Assertion | Action | MockCall | RawCode)[]
├── modifiers: Modifier[]                 // skip, only, timeout, retry
├── parameters: ParameterSet | null       // for parameterized tests
└── isAsync: boolean

Hook
├── type: 'beforeAll' | 'afterAll' | 'beforeEach' | 'afterEach' | 'around'
├── scope: 'suite' | 'module' | 'session'   // for pytest fixture scopes
├── body: (Action | MockCall | RawCode)[]
└── isAsync: boolean

Assertion
├── kind: AssertionKind                   // see enumeration below
├── subject: string                       // the expression being asserted on
├── expected: string | null               // expected value (if applicable)
├── isNegated: boolean
└── message: string | null                // custom failure message

MockCall
├── kind: 'createMock' | 'createSpy' | 'stubReturn' | 'stubImplementation'
│       | 'mockModule' | 'mockPartial' | 'restoreAll' | 'fakeTimers'
│       | 'fakeDate' | 'spyOnMethod'
├── target: string                        // what's being mocked
├── args: string[]                        // arguments to the mock call
└── returnValue: string | null

ImportStatement
├── kind: 'framework' | 'library' | 'relative' | 'sideEffect' | 'typeOnly'
├── source: string                        // module path
├── specifiers: { name: string, alias: string | null }[]
├── isDefault: boolean
└── isTypeOnly: boolean                   // TypeScript type-only imports

SharedVariable
├── name: string
├── initializer: string                   // the code that produces the value
├── isLazy: boolean                       // RSpec let (lazy) vs let! (eager)
└── scope: 'instance' | 'class' | 'module'

Modifier
├── type: 'skip' | 'only' | 'timeout' | 'retry' | 'tag' | 'pending'
│       | 'expectedFailure' | 'conditional'
├── value: string | number | null         // timeout ms, tag name, etc.
└── condition: string | null              // for conditional skip

ParameterSet
├── kind: 'values' | 'csv' | 'methodSource' | 'cartesian'
├── parameters: { name: string, values: string[] }[]
└── ids: string[] | null                  // named test case IDs

RawCode
├── code: string                          // passed through verbatim
├── comment: string | null                // HAMLET-TODO marker if unconvertible
└── confidence: number                    // 0.0–1.0 for this fragment

Comment
├── text: string
├── kind: 'inline' | 'block' | 'docstring' | 'directive' | 'license'
└── preserveExact: boolean                // true for license headers, directives
```

### AssertionKind Enumeration

```
equal, deepEqual, strictEqual, notEqual
truthy, falsy, isNull, isNotNull, isUndefined
isTrue, isFalse
instanceOf, typeOf
contains, notContains
matches (regex)
hasLength, hasCount
hasProperty, hasPropertyValue
hasAttribute, hasClass, hasCSS
greaterThan, lessThan, greaterOrEqual, lessOrEqual
closeTo (floating point)
throws, throwsMessage, doesNotThrow
rejects, resolves
isVisible, isHidden, isEnabled, isDisabled, isChecked
hasText, containsText, hasValue
calledWith, calledTimes, called
snapshot (flagged as framework-specific)
```

### Source Map

Every IR node carries a reference to its original source line range via the `sourceMap` on the root `TestFile` node. This enables:
- Confidence scoring per line
- Error reporting with source locations
- Side-by-side diff display

---

## 3. Framework Taxonomy

### By Paradigm

**BDD (describe/it/expect)**
- JavaScript: Jest, Vitest, Mocha+Chai, Jasmine
- Ruby: RSpec
- JavaScript E2E: Cypress, Playwright, WebdriverIO

**xUnit (class/method/assert)**
- Python: unittest
- Ruby: Minitest (test class mode)
- Java: JUnit 4, JUnit 5, TestNG
- JavaScript E2E: Selenium (test structure varies)

**Function-based (def test_/assert)**
- Python: pytest
- Ruby: Minitest (spec mode via `Minitest::Spec`)

**Imperative/procedural (no built-in test structure)**
- Puppeteer (uses any assertion library; test structure from Mocha/Jest)

### Conversion Difficulty by Direction

| Direction | Difficulty | Why |
|-----------|-----------|-----|
| Jest ↔ Vitest | Easy | Near-identical API. Namespace swap (`jest.*` → `vi.*`), import changes. |
| Jasmine → Jest | Easy | `spyOn`/`createSpy` syntax swap. Structure identical. |
| Jest → Jasmine | Easy | Reverse of above. `jest.fn` → `jasmine.createSpy`. |
| Mocha+Chai → Jest | Medium | Structure identical (`describe`/`it`). Chai assertion chains → Jest `expect`. Sinon → `jest.fn`/`jest.spyOn`. |
| Jest → Mocha+Chai | Medium | Reverse. Jest `expect` → Chai chains. `jest.fn` → Sinon stubs. |
| Puppeteer → Playwright | Low | Very similar page API. `page.$` → `page.locator`. Add Playwright `expect`. |
| Playwright → Puppeteer | Low | Reverse. Remove Playwright `expect`, add assertion library import. |
| WebdriverIO → Playwright | Medium | `$()` selectors → `page.locator()`. Sync/async mode differences. |
| WebdriverIO → Cypress | Medium | `$()` → `cy.get()`. Different chaining model. |
| Cypress ↔ Playwright | Done | Existing converters. Reference implementation. |
| Cypress ↔ Selenium | Done | Existing converters. |
| Playwright ↔ Selenium | Done | Existing converters. |
| pytest → unittest | Hard | Function-based → class-based. `assert` → `self.assertEqual`. Fixtures → `setUp`/`tearDown`. `conftest.py` hierarchy → class inheritance. `@pytest.mark.parametrize` → `self.subTest`. |
| unittest → pytest | Hard | Reverse. Strip classes. `self.assertEqual` → `assert`. `setUp` → fixtures. |
| nose2 → pytest | Medium | Similar to unittest with plugin layer. Class-based → function-based. |
| RSpec → Minitest | Hard | BDD → xUnit. `let`/`subject` → instance variables. `expect().to` → `assert_*`. Matchers → assertions. `shared_examples` → modules (lossy). Argument order flips (`expect(x).to eq(y)` → `assert_equal(y, x)`). |
| Minitest → RSpec | Hard | Reverse. `assert_*` → `expect().to`. Instance vars → `let`. |
| JUnit 4 → JUnit 5 | Medium | Annotation renames (`@Before` → `@BeforeEach`). `@Test(expected=...)` → `assertThrows` lambda. `@Rule` → `@ExtendWith`. Message parameter moves from first to last in assertion calls. |
| JUnit 5 ↔ TestNG | Medium | Different annotation names. `@ParameterizedTest` ↔ `@DataProvider`. TestNG `dependsOnMethods` is unconvertible to JUnit. |

### Mocking Libraries (Transform Plugins, Not Frameworks)

These pair with test frameworks and are handled as additional pattern layers:

| Library | Pairs With | Converts To |
|---------|-----------|-------------|
| Sinon.js | Mocha, Jasmine | Jest mocks, Vitest mocks |
| Chai | Mocha | Jest expect, Vitest expect |
| Mockito | JUnit 4, JUnit 5, TestNG | (stays as Mockito — mocking library is orthogonal) |
| unittest.mock | unittest | pytest-mock / `monkeypatch` |
| pytest-mock | pytest | unittest.mock |

---

## 4. Directory Structure

```
src/
  core/
    ir.js                        # IR node types, builder functions, visitor
    UniversalConverter.js         # Generic pipeline: parse → IR → emit
    FrameworkRegistry.js          # Framework registration, metadata, lookup
    ConfidenceScorer.js           # Per-file conversion confidence scoring
    BaseConverter.js              # Existing (kept for backward compat during migration)
    ConverterFactory.js           # Enhanced: routes all 30 directions
    PatternEngine.js              # Existing (unchanged)
    FrameworkDetector.js          # Enhanced: all 5 languages
    index.js                     # Barrel export
  languages/
    javascript/
      parser.js                  # JS/TS structural parser (tree-sitter or regex)
      frameworks/
        jest.js                  # Framework definition (see format below)
        vitest.js
        mocha.js
        jasmine.js
        cypress.js               # Migrated from existing patterns + converters
        playwright.js            # Migrated from existing patterns + converters
        selenium.js              # Migrated from existing patterns + converters
        webdriverio.js
        puppeteer.js
      detection.js               # JS-specific framework detection heuristics
      index.js                   # Registers all JS frameworks
    python/
      parser.js                  # Python structural parser
      frameworks/
        pytest.js
        unittest.js
        nose2.js
      detection.js
      index.js
    ruby/
      parser.js                  # Ruby structural parser
      frameworks/
        rspec.js
        minitest.js
      detection.js
      index.js
    java/
      parser.js                  # Java structural parser
      frameworks/
        junit4.js
        junit5.js
        testng.js
      detection.js
      index.js
  converters/                    # Existing 6 converters (legacy fallback during migration)
    CypressToPlaywright.js
    CypressToSelenium.js
    PlaywrightToCypress.js
    PlaywrightToSelenium.js
    SeleniumToCypress.js
    SeleniumToPlaywright.js
    index.js
  patterns/                      # Existing pattern files (referenced by migrated framework defs)
    commands/
      assertions.js
      interactions.js
      navigation.js
      selectors.js
      waits.js
  converter/                     # Existing utilities (unchanged)
    batchProcessor.js
    dependencyAnalyzer.js
    mapper.js
    metadataCollector.js
    orchestrator.js
    plugins.js
    repoConverter.js
    typescript.js
    validator.js
    visual.js
  utils/                         # Existing (unchanged)
    helpers.js
    reporter.js
  types/
    index.d.ts                   # Updated with new framework types
  index.js                       # Main API (enhanced)
```

---

## 5. Framework Definition File Format

Each framework definition exports a standardized object that the `FrameworkRegistry` consumes. This replaces the current approach where each converter class hardcodes its own patterns.

```javascript
// src/languages/javascript/frameworks/jest.js

export default {
  // Identity
  name: 'jest',
  language: 'javascript',
  paradigm: 'bdd',              // 'bdd' | 'xunit' | 'function'
  category: 'unit',             // 'unit' | 'e2e'

  // Detection heuristics (consumed by FrameworkDetector)
  detection: {
    imports: [
      /from\s+['"]@jest\/globals['"]/,
      /require\s*\(\s*['"]@jest\/globals['"]\s*\)/
    ],
    commands: [
      /jest\.fn\(/g,
      /jest\.mock\(/g,
      /jest\.spyOn\(/g
    ],
    filePatterns: [
      /\.test\.(js|ts|jsx|tsx)$/,
      /\.spec\.(js|ts|jsx|tsx)$/,
      /__tests__\//
    ],
    configFiles: [
      /jest\.config\.(js|ts|mjs|cjs)$/
    ],
    keywords: [
      /expect\([^)]+\)\.toBe\(/,
      /expect\([^)]+\)\.toEqual\(/,
      /describe\(/,
      /it\(/,
      /beforeEach\(/
    ],
    weights: { imports: 10, commands: 3, keywords: 2 }
  },

  // Patterns for PatternEngine (API-level transforms)
  patterns: {
    assertions: {
      'expect\\(([^)]+)\\)\\.toBe\\(': {
        ir: 'ASSERT_EQUAL',       // maps to IR AssertionKind
        args: ['$1']
      },
      'expect\\(([^)]+)\\)\\.toEqual\\(': {
        ir: 'ASSERT_DEEP_EQUAL',
        args: ['$1']
      },
      // ... full assertion pattern map
    },
    mocking: {
      'jest\\.fn\\(': { ir: 'MOCK_FN' },
      'jest\\.mock\\(': { ir: 'MOCK_MODULE' },
      'jest\\.spyOn\\(': { ir: 'SPY_ON' },
      // ...
    },
    async: {
      'jest\\.useFakeTimers\\(\\)': { ir: 'FAKE_TIMERS' },
      'jest\\.advanceTimersByTime\\(': { ir: 'ADVANCE_TIMERS' },
      // ...
    }
  },

  // Generators: IR node → target framework code
  generators: {
    ASSERT_EQUAL:       (subject, expected) => `expect(${subject}).toBe(${expected})`,
    ASSERT_DEEP_EQUAL:  (subject, expected) => `expect(${subject}).toEqual(${expected})`,
    ASSERT_TRUTHY:      (subject) => `expect(${subject}).toBeTruthy()`,
    ASSERT_THROWS:      (fn, errorType) => `expect(${fn}).toThrow(${errorType || ''})`,
    MOCK_FN:            () => 'jest.fn()',
    MOCK_MODULE:        (path) => `jest.mock(${path})`,
    SPY_ON:             (obj, method) => `jest.spyOn(${obj}, ${method})`,
    FAKE_TIMERS:        () => 'jest.useFakeTimers()',
    SUITE:              (name, body) => `describe(${name}, () => {\n${body}\n})`,
    TEST:               (name, body, isAsync) =>
                          `it(${name}, ${isAsync ? 'async ' : ''}() => {\n${body}\n})`,
    BEFORE_EACH:        (body, isAsync) =>
                          `beforeEach(${isAsync ? 'async ' : ''}() => {\n${body}\n})`,
    AFTER_EACH:         (body, isAsync) =>
                          `afterEach(${isAsync ? 'async ' : ''}() => {\n${body}\n})`,
    BEFORE_ALL:         (body, isAsync) =>
                          `beforeAll(${isAsync ? 'async ' : ''}() => {\n${body}\n})`,
    AFTER_ALL:          (body, isAsync) =>
                          `afterAll(${isAsync ? 'async ' : ''}() => {\n${body}\n})`,
    SKIP_TEST:          (name, body) => `it.skip(${name}, () => {\n${body}\n})`,
    ONLY_TEST:          (name, body) => `it.only(${name}, () => {\n${body}\n})`,
    PARAMETERIZED:      (name, params, body) =>
                          `it.each(${params})(${name}, ${body})`,
    // ...
  },

  // Import generation
  imports: {
    default: [],                 // Jest uses globals, no default import needed
    mocking: [],                 // Built-in
    assertions: [],              // Built-in
    explicit: [                  // When globals are disabled
      "import { describe, it, expect, jest } from '@jest/globals'"
    ]
  },

  // Structural templates
  structure: {
    wrapInSuite:     true,       // Uses describe() blocks
    wrapInClass:     false,      // No class wrapper needed
    wrapInFunction:  false,      // Tests are arrow functions in describe
    asyncPattern:    'async-arrow',  // async () => {}
    indentSize:      2,
    fileExtension:   '.test.js'
  }
};
```

---

## 6. Conversion Pipeline

> **Implementation status (v2.0):** The pipeline stages below are wired end-to-end
> in `ConversionPipeline` → `PipelineConverter`, which `ConverterFactory` uses for
> 20 of 25 conversion directions. However, current `emit()` functions in 13 of 15
> framework definitions receive the IR as `_ir` (unused) and perform regex-based
> string transforms on the raw source instead. The IR's primary runtime value today
> is **confidence scoring** via `ConfidenceScorer.score(ir)`. Emitters that
> reconstruct output from the IR tree are a planned improvement.
>
> Additionally, `src/converter/fileConverter.js` provides a standalone
> `convertFile()` / `convertCypressToPlaywright()` path that bypasses the pipeline
> entirely — it uses hardcoded regex patterns with no IR, no PatternEngine, and no
> ConverterFactory.

### UniversalConverter Flow

```
Source Code
    │
    ▼
┌──────────────┐
│  1. DETECT   │  FrameworkDetector identifies source framework + language
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  2. PARSE    │  Regex-based line classifier → IR (TestFile node tree)
│              │  - Classifies lines as suites, tests, hooks, assertions, raw code
│              │  - Preserves raw code for unrecognized constructs
│              │  - Preserves comments
│              │  (Planned: tree-sitter for block boundaries and nesting depth)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  3. TRANSFORM│  IR-to-IR structural transforms
│              │  - Paradigm shift (BDD → xUnit, function → class)
│              │  - Hook type mapping (before → setUp, fixture → beforeEach)
│              │  - Modifier conversion (skip, only, timeout, tags)
│              │  - Parameter conversion (parametrize → each/DataProvider)
│              │  - Shared state conversion (let → instance var)
│              │  (Currently lightweight — most transforms happen in emit via regex)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  4. EMIT     │  Regex-based string transforms on source code
│              │  - Detects source framework from content, applies pattern maps
│              │  - Rewrites imports, API calls, assertions, and test structure
│              │  - Marks unconvertible nodes with HAMLET-TODO comments
│              │  (Planned: reconstruct output from IR tree instead of source string)
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  5. SCORE    │  ConfidenceScorer walks the IR tree
│              │  - Counts converted vs unconvertible nodes
│              │  - Produces per-line annotations for anything below 100%
│              │  - Assigns overall file confidence percentage
└──────┬───────┘
       │
       ▼
Target Code + Confidence Report
```

### Short-Circuit for Same-Paradigm Conversions

When source and target share the same paradigm (e.g., Jest → Vitest), the pipeline can skip the IR structural transform step entirely and use PatternEngine-only conversion. This is faster and produces higher-fidelity output for trivial conversions.

```
Jest → Vitest:      DETECT → PARSE → EMIT (regex on source) → SCORE
pytest → unittest:  DETECT → PARSE → TRANSFORM → EMIT (regex on source) → SCORE
```

> **Implementation status (v2.0):** In practice, all 20 pipeline directions
> currently follow the same flow — parse produces IR for scoring, but emit
> operates on the source string via regex in every case. The distinction between
> "short-circuit" and "full pipeline" is in the TRANSFORM step (lightweight for
> same-paradigm, more involved for cross-paradigm), not in whether emit uses IR.

---

## 7. Parsing Strategy

> **Implementation status (v2.0):** tree-sitter is not yet integrated.
> All 15 framework parsers currently use regex-based line-by-line classification
> (see e.g. `src/languages/javascript/frameworks/cypress.js:parse()`).
> This works well for flat test files but does not track nesting depth or block
> boundaries. The tree-sitter plan below remains the intended direction for
> improving structural accuracy.

### Planned: Per-Language Parser Approach

| Language | Planned Strategy | Rationale |
|----------|----------|-----------|
| JavaScript/TypeScript | tree-sitter + regex | tree-sitter identifies `describe`/`it`/`test` block boundaries, class declarations, function signatures, and import statements. Regex handles API-level transforms within identified blocks. Tree-sitter's JS/TS grammar is mature and handles JSX/TSX. |
| Python | tree-sitter + regex | tree-sitter identifies `class` boundaries, `def` functions, decorators (`@pytest.mark.*`, `@unittest.skip`), and `import` statements. Critical for pytest → unittest where functions must be wrapped in classes. Python's significant whitespace makes pure regex unreliable for structural parsing. |
| Ruby | tree-sitter + regex | tree-sitter identifies `describe`/`context`/`it` blocks, `class` definitions, `def` methods, and `do...end`/`{ }` block boundaries. Necessary for RSpec → Minitest where BDD blocks become class methods. |
| Java | tree-sitter + regex | tree-sitter identifies class declarations, method declarations, annotations (including their arguments), and import statements. Required for JUnit 4 → 5 where `@Test(expected=X.class)` must wrap the method body in `assertThrows`. |

### Why tree-sitter (planned)

- Single dependency that handles all 4 languages from Node.js (no Python/Ruby/Java runtime required)
- Incremental parsing — can parse partial/broken files gracefully
- Concrete syntax tree preserves comments, whitespace, and formatting
- Battle-tested grammars maintained by the tree-sitter community
- npm packages: `tree-sitter`, `tree-sitter-javascript`, `tree-sitter-python`, `tree-sitter-ruby`, `tree-sitter-java`
- Alternative: `web-tree-sitter` (WASM-based, no native compilation needed)

### Current: Regex-Only Parsing

All framework `parse()` functions currently iterate source lines and classify each
line into an IR node type using regex tests (e.g. `/\bdescribe\s*\(/` → `TestSuite`,
`/\bit\s*\(/` → `TestCase`). This produces a flat IR — one node per line with no
nesting information. The IR is consumed by `ConfidenceScorer` for scoring but is
not used by emitters for code generation.

### Planned: What tree-sitter Would Add vs What Regex Does

**tree-sitter would identify (not yet implemented):**
- Block boundaries (where does this `describe` block start and end?)
- Nesting depth (is this `it()` inside a `describe` or at top level?)
- Function/method signatures (parameters, async keyword, decorators/annotations)
- Import statements (full statement boundaries)
- Class structure (which methods belong to which class?)
- Comments (attached to the correct node)

**Regex transforms (current implementation):**
- API calls (`cy.get('.btn')` → `page.locator('.btn')`)
- Assertion syntax (`expect(x).toBe(y)` → `assert x == y`)
- Mock calls (`jest.fn()` → `vi.fn()`)
- Command mappings (`.type(text)` → `.fill(text)`)
- Import path rewrites (`from 'cypress'` → `from '@playwright/test'`)

---

## 8. Confidence Scoring System

### Algorithm

```
FileConfidence = (convertedWeight / totalWeight) × 100

Where:
  totalWeight     = sum of weights for all recognized patterns in the source file
  convertedWeight = sum of weights for patterns successfully converted

Weight assignments:
  Structural node (suite, test, hook)       = 3
  Assertion                                 = 2
  Mock/spy call                             = 2
  API command (navigation, selector, etc.)  = 1
  Import statement                          = 1
  Raw code (passed through)                 = 0  (neutral — doesn't count)
  Unconvertible (HAMLET-TODO)               = 0  (counts against totalWeight)
```

### Per-File Report Format

```
┌─────────────────────────────────────────────────┐
│ Conversion: jest → vitest                       │
│ File: auth.test.js                              │
│ Confidence: 94%                                 │
│                                                 │
│ Converted:      47/50 patterns                  │
│ Unconvertible:  2 (HAMLET-TODO markers added)   │
│ Warnings:       1 (may need manual review)      │
│                                                 │
│ Issues:                                         │
│   Line 23: jest.mock hoisting → vi.mock (verify │
│            factory function behavior)            │
│   Line 45: __mocks__ dir convention (move to    │
│            __mocks__ or inline vi.mock)          │
│   Line 89: Custom snapshot serializer (no       │
│            Vitest equivalent)                    │
└─────────────────────────────────────────────────┘
```

### Confidence Thresholds

| Range | Label | Meaning |
|-------|-------|---------|
| 90–100% | High | Output is likely correct. Review recommended but not required. |
| 70–89% | Medium | Some patterns need manual review. HAMLET-TODO markers present. |
| Below 70% | Low | Significant manual work needed. Structural conversion may be incomplete. |

---

## 9. Unconvertible Pattern Output Format

When the converter encounters code it cannot translate, it produces a structured TODO comment in the target language's comment syntax. Output must never silently drop code or produce syntactically broken output.

### Format

```
// HAMLET-TODO [UNCONVERTIBLE-XXX]: <description>
// Original: <original source line(s)>
// Manual action required: <what the developer needs to do>
<best-effort converted code OR original code as comment>
```

### Examples

**JavaScript:**
```javascript
// HAMLET-TODO [UNCONVERTIBLE-001]: Jest snapshot testing has no Vitest equivalent
//   for custom serializers.
// Original: expect(tree).toMatchSnapshot({ snapshotSerializers: [mySerializer] })
// Manual action required: Configure Vitest snapshot serializer in vitest.config.ts
//   or replace with explicit assertion.
expect(tree).toMatchSnapshot();
```

**Python:**
```python
# HAMLET-TODO [UNCONVERTIBLE-003]: pytest conftest.py fixture hierarchy
#   has no unittest equivalent.
# Original: @pytest.fixture(scope="session")
# Manual action required: Move fixture logic to setUpClass or a base test class.
```

**Ruby:**
```ruby
# HAMLET-TODO [UNCONVERTIBLE-004]: RSpec shared_examples cannot be
#   automatically converted to Minitest.
# Original: it_behaves_like 'a searchable resource'
# Manual action required: Extract shared example tests into a module
#   or helper methods and include them manually.
```

**Java:**
```java
// HAMLET-TODO [UNCONVERTIBLE-006]: JUnit 4 @Rule requires manual
//   conversion to JUnit 5 @ExtendWith or @RegisterExtension.
// Original: @Rule public ExpectedException thrown = ExpectedException.none();
// Manual action required: Replace with assertThrows() calls at each usage site.
```

### Error Recovery Policy

Default behavior: **skip and warn**. Unrecognized code passes through unchanged with a `HAMLET-WARNING` comment above it. The file still converts — it just has lower confidence.

CLI flag to override:
```
--on-error skip       # Default. Pass through unrecognized code unchanged.
--on-error fail       # Abort file conversion on first unrecognized pattern.
--on-error best-effort  # Attempt conversion even for uncertain patterns (lower confidence).
```

---

## 10. Import Rewriting Rules

Every conversion rewrites framework imports. This is always the first change applied (before structural transforms) and the last verified (after all transforms).

### JavaScript

```
// Jest → Vitest
// Input (Jest uses globals, no import)
// Output:
import { describe, it, expect, vi } from 'vitest';
// All jest.* → vi.*

// Vitest → Jest
// Input:
import { describe, it, expect, vi } from 'vitest';
// Output (remove import, Jest uses globals):
// (no import needed — Jest injects globals)
// All vi.* → jest.*

// Mocha+Chai → Jest
// Input:
const { expect } = require('chai');
const sinon = require('sinon');
// Output:
// (remove both — Jest provides expect globally, jest.fn replaces sinon)

// Cypress → Playwright
// Input (Cypress uses globals):
// Output:
import { test, expect } from '@playwright/test';
```

### Python

```
# pytest → unittest
# Input:
import pytest
# Output:
import unittest

# unittest → pytest
# Input:
import unittest
# Output:
import pytest
```

### Ruby

```
# RSpec → Minitest
# Input:
require 'spec_helper'
# Output:
require 'minitest/autorun'

# Minitest → RSpec
# Input:
require 'minitest/autorun'
# Output:
require 'spec_helper'
```

### Java

```
// JUnit 4 → JUnit 5
// Input:
import org.junit.Test;
import org.junit.Before;
import org.junit.After;
import org.junit.BeforeClass;
import org.junit.AfterClass;
import org.junit.Ignore;
import org.junit.Assert.*;
import org.junit.Rule;
import org.junit.runner.RunWith;
// Output:
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Assertions.*;
// @Rule and @RunWith imports removed (refactored to extensions)
```

---

## 11. Output Formatting Rules

Converted code matches the idiomatic style of the target framework and language.

| Property | JavaScript | Python | Ruby | Java |
|----------|-----------|--------|------|------|
| Indent | Preserve source (2 or 4 space), default 2 | 4 spaces (PEP 8) | 2 spaces | 4 spaces |
| Naming | camelCase | snake_case | snake_case | camelCase |
| Blank lines between tests | 1 | 2 (top-level), 1 (methods) | 1 | 1 |
| Trailing whitespace | Stripped | Stripped | Stripped | Stripped |
| Final newline | Exactly one | Exactly one | Exactly one | Exactly one |
| Line wrapping | Don't introduce new breaks | Don't introduce new breaks | Don't introduce new breaks | Don't introduce new breaks |
| Quotes | Preserve source style | Preserve source style | Preserve source style | N/A (Java uses `"` only) |

---

## 12. Comment & Documentation Preservation

Comments are first-class citizens in the IR. The parser attaches each comment to its nearest structural node. The emitter reproduces them in the target syntax.

| Comment Type | Behavior |
|-------------|----------|
| Inline (`// ...`, `# ...`) within test body | Preserved verbatim |
| Block (`/* ... */`, `=begin...=end`, `""" """`) above test | Preserved, converted to target comment syntax |
| JSDoc / docstring on test function | Preserved, converted to target docstring syntax |
| TODO / FIXME / HACK | Preserved verbatim |
| Commented-out code | Preserved with `HAMLET-WARNING: commented-out code` |
| Directives (`// eslint-disable`, `# type: ignore`, `# noinspection`) | Preserved exactly (framework-agnostic) |
| License headers | Preserved exactly at top of file |
| Region markers (`// #region`, `// MARK:`) | Preserved exactly |

---

## 13. TypeScript Handling

TypeScript test files are handled by the same JavaScript parser and PatternEngine. Type-specific constructs are preserved through conversion:

| Pattern | Handling |
|---------|----------|
| Type annotations on variables | Passed through as `RawCode` (preserved) |
| Type assertions (`as Type`, `<Type>`) | Passed through (preserved) |
| `import type { Foo }` | Preserved as `ImportStatement` with `isTypeOnly: true` |
| Generic test helpers | Passed through (preserved) |
| `jest.fn<ReturnType>()` → `vi.fn<ReturnType>()` | Handled by PatternEngine (type parameter preserved) |
| `as const` assertions | Passed through (preserved) |
| `.d.ts` test utility files | Passed through unchanged (no test structure to convert) |

The principle: type annotations don't affect test behavior. They are preserved verbatim through conversion. Only the runtime test API calls are transformed.

---

## 14. Configuration File Conversion

Config file conversion is handled separately from test file conversion. Each framework definition includes a `configMapping` that maps source config keys to target config keys.

### Supported Config Conversions

| Source | Target | Source File | Target File |
|--------|--------|-------------|-------------|
| Jest | Vitest | `jest.config.js` | `vitest.config.ts` |
| Cypress | Playwright | `cypress.config.js` | `playwright.config.ts` |
| WebdriverIO | Playwright | `wdio.conf.js` | `playwright.config.ts` |
| pytest | unittest | `pytest.ini` / `pyproject.toml` | N/A (unittest uses CLI args) |
| RSpec | Minitest | `.rspec` | Rakefile config |
| JUnit 4 | JUnit 5 | `pom.xml` / `build.gradle` (deps only) | Updated deps |
| TestNG | JUnit 5 | `testng.xml` | `@Suite` + `@SelectClasses` |

### Key Config Mappings (Jest → Vitest Example)

```
jest.config.js transform          → vitest plugins / esbuild
jest.config.js moduleNameMapper   → vitest resolve.alias
jest.config.js testEnvironment    → vitest environment
jest.config.js setupFiles         → vitest setupFiles
jest.config.js globals            → vitest globals: true
jest.config.js testMatch          → vitest include
jest.config.js coverageThreshold  → vitest coverage.thresholds
```

Config conversion confidence is reported separately from test file confidence, since config files are inherently more lossy (framework-specific options often have no equivalent).

---

## 15. CLI Specification

### New Framework Detection

```bash
hamlet detect ./tests/test_auth.py
# Output: pytest (confidence: 92%), unittest (confidence: 8%)

hamlet detect ./tests/AuthTest.java
# Output: JUnit 4 (confidence: 88%), TestNG (confidence: 12%)

hamlet detect ./tests/auth.test.js
# Output: Jest (confidence: 95%), Mocha (confidence: 5%)
```

### Conversion Commands

The existing `--from`/`--to` syntax extends to all frameworks:

```bash
# Existing (unchanged)
hamlet convert ./tests/ --from cypress --to playwright -o ./output/

# New directions
hamlet convert ./tests/ --from jest --to vitest -o ./output/
hamlet convert ./tests/ --from pytest --to unittest -o ./output/
hamlet convert ./tests/ --from rspec --to minitest -o ./output/
hamlet convert ./tests/ --from junit4 --to junit5 -o ./output/
hamlet convert ./tests/ --from mocha --to jest -o ./output/
```

### Shorthand Commands

```bash
hamlet jest2vt <source>       # Jest → Vitest
hamlet vt2jest <source>       # Vitest → Jest
hamlet mocha2jest <source>    # Mocha+Chai → Jest
hamlet j4toj5 <source>        # JUnit 4 → JUnit 5
hamlet pytest2ut <source>     # pytest → unittest
hamlet ut2pytest <source>     # unittest → pytest
hamlet rspec2mini <source>    # RSpec → Minitest
hamlet mini2rspec <source>    # Minitest → RSpec
hamlet pup2pw <source>        # Puppeteer → Playwright
hamlet wdio2pw <source>       # WebdriverIO → Playwright
```

### Batch Mode

Mixed-framework directories are handled gracefully:

```bash
hamlet convert ./tests/ --from jest --to vitest -o ./output/
# Skips non-Jest files with warning, does not fail:
#   SKIP: helper.js (not a test file)
#   SKIP: cypress/e2e/login.cy.js (Cypress, not Jest)
#   CONVERT: auth.test.js → auth.test.js (confidence: 96%)
#   CONVERT: api.test.js → api.test.js (confidence: 82%)
```

### Dry Run with Confidence

```bash
hamlet convert ./tests/ --from jest --to vitest --dry-run
# Output:
#   auth.test.js        → auth.test.js        (confidence: 96%)
#   api.test.js         → api.test.js         (confidence: 82%)
#   legacy.test.js      → legacy.test.js      (confidence: 61%) ⚠
#   snapshot.test.js    → snapshot.test.js     (confidence: 34%) ⚠⚠
```

---

## 16. Dependency Policy

### Required Dependencies (New)

| Package | Purpose | Size |
|---------|---------|------|
| `web-tree-sitter` | WASM-based tree-sitter runtime (no native compilation) | ~300KB |
| `tree-sitter-javascript` | JS/TS grammar for structural parsing | ~200KB |
| `tree-sitter-python` | Python grammar | ~150KB |
| `tree-sitter-ruby` | Ruby grammar | ~150KB |
| `tree-sitter-java` | Java grammar | ~150KB |

Total: ~950KB added to dependencies. All are WASM-based (via `web-tree-sitter`), so no native compilation or language runtimes are required. Users do not need Python, Ruby, or Java installed to convert those languages.

### Rejected Alternatives

- **Shelling out to language runtimes** (Python `ast` module, Ruby `parser` gem): Requires users to have those languages installed. Breaks the "Node.js CLI tool" constraint.
- **Babel/acorn for JavaScript**: Only handles JS/TS. Would need separate parsers for other languages anyway.
- **No parser (regex-only)**: Insufficient for cross-paradigm structural transforms. Cannot reliably identify block boundaries in Python (significant whitespace) or Ruby (multi-line blocks).

---

## 17. Migration Plan for Existing Converters

All 207 existing tests must pass throughout migration. The migration has 3 phases.

### Phase 1: Extract Framework Definitions

Extract patterns from the 6 existing converter classes and 5 pattern files into the new framework definition format.

**From:**
- `src/converters/CypressToPlaywright.js` (638 lines) → patterns split into `cypress.js` and `playwright.js` framework definitions
- `src/converters/CypressToSelenium.js` (392 lines) → patterns added to `cypress.js` and `selenium.js`
- `src/converters/PlaywrightToCypress.js` (474 lines) → patterns added to `playwright.js` and `cypress.js`
- `src/converters/PlaywrightToSelenium.js` (382 lines) → patterns added to `playwright.js` and `selenium.js`
- `src/converters/SeleniumToCypress.js` (335 lines) → patterns added to `selenium.js` and `cypress.js`
- `src/converters/SeleniumToPlaywright.js` (396 lines) → patterns added to `selenium.js` and `playwright.js`
- `src/patterns/commands/*.js` (5 files) → `directMappings` absorbed into framework definitions

**To:**
- `src/languages/javascript/frameworks/cypress.js`
- `src/languages/javascript/frameworks/playwright.js`
- `src/languages/javascript/frameworks/selenium.js`

**Verification:** Existing 207 tests pass using legacy converters (no behavior change yet).

### Phase 2: Route Through UniversalConverter

Register the 3 E2E frameworks in `FrameworkRegistry`. Update `UniversalConverter` to handle Cypress ↔ Playwright ↔ Selenium conversions using the new pipeline.

Update `ConverterFactory` to route through `UniversalConverter` first, falling back to legacy converters if the new path fails:

```javascript
async createConverter(from, to, options) {
  try {
    return new UniversalConverter(from, to, options);  // new path
  } catch (e) {
    return this.createLegacyConverter(from, to, options);  // fallback
  }
}
```

**Verification:** All 207 tests pass through the new `UniversalConverter` path. Legacy converters are still present but unused.

### Phase 3: Remove Legacy Converters

Once all tests pass through `UniversalConverter`:
1. Remove the fallback in `ConverterFactory`
2. Delete `src/converters/CypressToPlaywright.js` and the other 5 converter files
3. Keep `src/converters/index.js` as a re-export pointing to `UniversalConverter` for backward compatibility of external imports

**Verification:** All 207 tests still pass. `src/converters/` directory is either empty or contains only the backward-compat re-export.

---

## 18. Difficulty Assessment

| Conversion | Difficulty | Key Challenges |
|-----------|-----------|----------------|
| Jest ↔ Vitest | Easy | Near-identical APIs. Namespace swap (`jest.*` → `vi.*`). `jest.mock` hoisting semantics differ subtly from `vi.mock`. |
| Jasmine → Jest | Easy | `spyOn`/`createSpy` syntax swap. `jasmine.createSpyObj` → multiple `jest.fn()`. `jasmine.clock()` → `jest.useFakeTimers()`. |
| Mocha+Chai → Jest | Medium | Chai assertion chains (`.to.be.a('string')`, `.deep.equal`, `.have.lengthOf`) have no 1:1 Jest equivalent — each chain must be decomposed. Sinon stubs/spies → Jest mocks. `done()` callback → `async/await`. |
| Puppeteer → Playwright | Low | Very similar page API. `page.$()` → `page.locator()`. Puppeteer has no built-in assertions — must detect which assertion library is used (Jest, Chai, Node assert) and convert that too. |
| WebdriverIO → Playwright | Medium | `$()` selectors → `page.locator()`. Sync mode (WDIO v7) vs async mode (WDIO v8) — must detect which. `browser.mock()` → `page.route()`. Custom commands → Playwright fixtures. |
| pytest → unittest | Hard | Function-based → class-based. `assert x == y` → `self.assertEqual(x, y)` (15+ assertion mappings). `@pytest.fixture` → `setUp`/`tearDown` (scope mapping). `conftest.py` → base class inheritance. `@pytest.mark.parametrize` → `self.subTest()`. Fixture dependency chains. |
| unittest → pytest | Hard | Reverse of above. Strip class wrappers. `self.assert*` → bare `assert`. `setUp`/`tearDown` → fixtures. |
| nose2 → pytest | Medium | Largely unittest-compatible. `@such.it` DSL → pytest functions. `params` decorator → `@pytest.mark.parametrize`. Layer-based setup → fixtures. |
| RSpec → Minitest | Hard | BDD → xUnit. `let`/`subject` → instance variables in `setup`. `expect().to eq()` → `assert_equal()` (argument order flips!). `shared_examples` → included modules (lossy). `allow().to receive()` → `Minitest::Mock`. `change { }.by()` has no equivalent. |
| Minitest → RSpec | Hard | Reverse. `assert_*` → `expect().to`. Instance vars → `let`. `setup` → `before(:each)`. |
| JUnit 4 → JUnit 5 | Medium | Annotation renames. `@Test(expected=X)` → `assertThrows(X, () -> { })` (requires wrapping method body in lambda). `@Rule` → `@ExtendWith` (each Rule type maps differently). Message parameter position changes in assertions. |
| JUnit 5 → TestNG | Medium | `@BeforeEach` → `@BeforeMethod`. `@ParameterizedTest` → `@DataProvider`. `@Nested` classes → flat classes (lossy). `Assertions.*` → `Assert.*`. |
| TestNG → JUnit 5 | Medium | `@BeforeMethod` → `@BeforeEach`. `@DataProvider` → `@MethodSource`. `dependsOnMethods` is unconvertible. `@Test(priority=N)` → `@Order(N)` (different semantics). |

---

## 19. Known Unconvertible Patterns

These patterns produce HAMLET-TODO comments rather than broken output. Each has a unique ID for tracking.

| ID | Pattern | Affected Directions | Why |
|----|---------|-------------------|-----|
| UNCONVERTIBLE-001 | Snapshot testing (Jest/Vitest `toMatchSnapshot`) | Jest/Vitest → any non-JS | Snapshot format is framework-specific. No equivalent in Python/Ruby/Java. |
| UNCONVERTIBLE-002 | Cypress custom commands (`Cypress.Commands.add`) | Cypress → any | Custom commands are project-specific. Must be manually refactored to target framework's extension mechanism. |
| UNCONVERTIBLE-003 | pytest `conftest.py` fixture hierarchy | pytest → unittest | No equivalent for directory-scoped, auto-discovered fixtures in unittest. |
| UNCONVERTIBLE-004 | RSpec `shared_examples` / `shared_context` | RSpec → Minitest | No declarative shared behavior mechanism. Must manually extract to modules. |
| UNCONVERTIBLE-005 | RSpec `let!` eager evaluation semantics | RSpec → Minitest | `let!` runs before each example regardless of usage. Instance vars in `setup` are the closest equivalent but semantically different. |
| UNCONVERTIBLE-006 | JUnit 4 `@Rule` (custom TestRule) | JUnit 4 → JUnit 5 | Each custom Rule must be individually refactored to implement `BeforeEachCallback`/`AfterEachCallback`. Automated conversion cannot know the Rule's semantics. |
| UNCONVERTIBLE-007 | TestNG `dependsOnMethods` | TestNG → JUnit 5 | JUnit tests are designed to be independent. No dependency ordering mechanism. |
| UNCONVERTIBLE-008 | TestNG XML suite configuration | TestNG → JUnit 5 | No equivalent for XML-based suite/group/parameter configuration. |
| UNCONVERTIBLE-009 | Playwright custom fixtures (`test.extend`) | Playwright → Cypress | Cypress has no fixture system. Must be refactored to custom commands or support files. |
| UNCONVERTIBLE-010 | Property-based tests | Any → any | Property-based → example-based is fundamentally lossy. |
| UNCONVERTIBLE-011 | Mocha TDD interface (`suite`/`test`) mixed with BDD | Mocha → Jest | Mocha allows mixing interfaces; Jest does not. |
| UNCONVERTIBLE-012 | WebdriverIO sync mode idioms | WDIO sync → async targets | Sync mode requires `@wdio/sync` which transforms the entire execution model. |
| UNCONVERTIBLE-013 | Puppeteer `page.$eval` | Puppeteer → Cypress | Different execution model (Puppeteer runs JS in browser context; Cypress chains commands). |
| UNCONVERTIBLE-014 | pytest `monkeypatch` | pytest → unittest | Maps partially to `unittest.mock.patch` but different API. Best-effort with warning. |
| UNCONVERTIBLE-015 | RSpec `change { }.by()` matcher | RSpec → Minitest | No equivalent. Must manually assert before/after values. |
| UNCONVERTIBLE-016 | Chai plugin assertions | Mocha+Chai → Jest | Per-plugin mapping needed. Generic plugins cannot be auto-converted. |
| UNCONVERTIBLE-017 | TestNG parallel thread config | TestNG → JUnit 5 | Different parallel execution model. |
| UNCONVERTIBLE-018 | Custom matchers/assertions (any framework) | Any → any | Project-specific. Cannot auto-convert without understanding the matcher's semantics. |
| UNCONVERTIBLE-019 | Framework-specific config options | Any → any | Lossy. Best-effort with per-key confidence. |
| UNCONVERTIBLE-020 | Implicit globals → explicit imports | Jasmine/Jest → Vitest | Vitest requires explicit imports. All global usage sites must be identified. Best-effort with verification warning. |

---

## 20. Test Strategy

### Test File Convention

Each test case has an input file, an expected output file, and a test runner file:

```
test/javascript/jest-to-vitest/mocking/MOCK-001.input.js     # Jest source
test/javascript/jest-to-vitest/mocking/MOCK-001.expected.js   # Expected Vitest output
test/javascript/jest-to-vitest/mocking/MOCK-001.test.js       # Test runner
```

The test runner reads input, runs conversion, and diffs against expected output:

```javascript
import { UniversalConverter } from '../../../../src/core/UniversalConverter.js';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

describe('MOCK-001: Function spy (track calls without replacing)', () => {
  it('converts jest.fn() to vi.fn()', async () => {
    const input = await fs.readFile(path.join(__dirname, 'MOCK-001.input.js'), 'utf8');
    const expected = await fs.readFile(path.join(__dirname, 'MOCK-001.expected.js'), 'utf8');
    const converter = new UniversalConverter('jest', 'vitest');
    const result = await converter.convert(input);
    expect(result.content).toBe(expected);
  });
});
```

### Test Organization

```
test/
├── e2e/
│   ├── cypress-to-playwright/
│   │   ├── structure/              # STRUCTURE-001–007
│   │   ├── hooks/                  # HOOKS-001–010
│   │   ├── assertions/             # E2E-VIS-001–014
│   │   ├── navigation/             # E2E-NAV-001–011
│   │   ├── selectors/              # E2E-SEL-001–011
│   │   ├── forms/                  # E2E-FORM-001–012
│   │   ├── waiting/                # E2E-WAIT-001–010
│   │   ├── network/                # E2E-NET-001–009
│   │   ├── state/                  # E2E-STATE-001–008
│   │   ├── cypress-specific/       # E2E-CY-001–020
│   │   └── unconvertible/          # Relevant UNCONVERTIBLE cases
│   ├── playwright-to-cypress/
│   ├── webdriverio-to-playwright/
│   ├── puppeteer-to-playwright/
│   └── ...                         # All 12 E2E directions
├── javascript/
│   ├── jest-to-vitest/
│   │   ├── structure/
│   │   ├── hooks/
│   │   ├── assertions/
│   │   ├── async/
│   │   ├── mocking/
│   │   ├── modifiers/
│   │   ├── parameterized/
│   │   ├── imports/
│   │   ├── jest-specific/          # JS-JEST-001–015
│   │   └── unconvertible/
│   ├── mocha-to-jest/
│   ├── jasmine-to-jest/
│   └── ...                         # All 10 JS unit directions
├── python/
│   ├── pytest-to-unittest/
│   │   ├── structure/
│   │   ├── assertions/             # PY-ASSERT-MAP-001–015
│   │   ├── fixtures/               # PY-PYTEST-003–006, 012–014
│   │   ├── parametrize/            # PARAM-001–010
│   │   ├── pytest-specific/        # PY-PYTEST-001–025
│   │   └── unconvertible/
│   ├── unittest-to-pytest/
│   └── nose2-to-pytest/
├── ruby/
│   ├── rspec-to-minitest/
│   │   ├── structure/
│   │   ├── assertions/             # RB-ASSERT-MAP-001–015
│   │   ├── rspec-specific/         # RB-RSPEC-001–025
│   │   └── unconvertible/
│   └── minitest-to-rspec/
├── java/
│   ├── junit4-to-junit5/
│   │   ├── structure/
│   │   ├── assertions/             # JAVA-ASSERT-MAP-001–025
│   │   ├── junit4-specific/        # JAVA-J4-001–025
│   │   └── unconvertible/
│   ├── junit5-to-testng/
│   └── testng-to-junit5/
├── config/                         # CONFIG-001–024
│   ├── jest-to-vitest/
│   ├── cypress-to-playwright/
│   └── ...
├── universal/                      # Cross-cutting categories
│   ├── comments/                   # COMMENT-001–010
│   ├── typescript/                 # TS-001–012
│   ├── test-data/                  # DATA-001–012
│   ├── values/                     # VALUE-001–015
│   ├── multi-file/                 # MULTI-001–010
│   ├── class-patterns/             # CLASS-001–012
│   ├── output-capture/             # OUTPUT-001–006
│   ├── cleanup/                    # CLEANUP-001–010
│   ├── decorator-stacking/         # STACK-001–008
│   └── concurrency/                # CONCURRENT-001–008
└── fixtures/                       # Shared test data
    ├── input/
    └── expected/
```

### Test Count Estimate

| Category | Cases | x Directions | Total |
|----------|-------|-------------|-------|
| Universal (STRUCTURE through IMPORT) | ~100 | 30 | ~3,000 |
| E2E-specific (NAV, SEL, FORM, WAIT, NET, STATE, VIS) | ~82 | 12 | ~984 |
| Framework-specific (JS-JEST, PY-PYTEST, etc.) | ~325 | 1 each | ~325 |
| Cross-framework assertion mappings | ~90 | 1 each | ~90 |
| Config conversion | ~24 | 1 each | ~24 |
| Unconvertible verification | ~20 | varies | ~100 |
| New universal categories (comments, TS, data, values, multi-file, class, output, cleanup, decorators, concurrency) | ~91 | varies | ~500 |
| **Total (before pruning)** | | | **~5,023** |

Not every universal test applies to every direction (e.g., MOCK-005 module mocking doesn't apply identically to Java). After pruning inapplicable cases:

**Realistic target: 3,500–4,500 tests**

Each test is a small fixture pair (input file + expected output), so the suite runs fast despite the count.

---

## 21. Implementation Sequence

### Step 1: Core Infrastructure

Build the foundation without any new converters:
1. `ir.js` — IR node types and builder
2. `FrameworkRegistry.js` — Registration, lookup, metadata
3. `UniversalConverter.js` — Parse → IR → emit pipeline
4. `ConfidenceScorer.js` — Per-file scoring
5. Language parsers (tree-sitter integration)
6. Migrate existing Cypress/Playwright/Selenium into framework definitions
7. All 207 existing tests pass through the new pipeline

### Step 2: JavaScript Unit Testing (Easiest)

Test fixtures first (red suite), then implementation:
1. Jest ↔ Vitest (easiest — validates the pipeline works for trivial conversions)
2. Mocha+Chai → Jest (validates Chai assertion chain decomposition)
3. Jasmine → Jest (validates spy/mock mapping)
4. Reverse directions

### Step 3: Java Unit Testing

1. JUnit 4 → JUnit 5 (validates annotation refactoring + lambda wrapping)
2. JUnit 5 ↔ TestNG (validates cross-framework annotation mapping)

### Step 4: Python Unit Testing

1. pytest → unittest (validates paradigm shift: function → class)
2. unittest → pytest (validates class stripping)
3. nose2 → pytest

### Step 5: Ruby Unit Testing

1. RSpec → Minitest (validates BDD → xUnit paradigm shift)
2. Minitest → RSpec

### Step 6: E2E Expansion

1. Puppeteer → Playwright (easy — validates E2E expansion works)
2. WebdriverIO ↔ Playwright
3. WebdriverIO ↔ Cypress
4. Playwright → Puppeteer

### Step 7: Configuration Conversion

Config conversion for all supported directions.

---

## 22. Framework Type Definition Update

The `Framework` type in `src/types/index.d.ts` expands from:

```typescript
type Framework = 'cypress' | 'playwright' | 'selenium';
```

To:

```typescript
type Framework =
  // E2E
  | 'cypress' | 'playwright' | 'selenium' | 'webdriverio' | 'puppeteer'
  // JavaScript unit
  | 'jest' | 'vitest' | 'mocha' | 'jasmine'
  // Python unit
  | 'pytest' | 'unittest' | 'nose2'
  // Ruby unit
  | 'rspec' | 'minitest'
  // Java unit
  | 'junit4' | 'junit5' | 'testng';

type Language = 'javascript' | 'python' | 'ruby' | 'java';

type Paradigm = 'bdd' | 'xunit' | 'function';

type Category = 'unit' | 'e2e';
```
