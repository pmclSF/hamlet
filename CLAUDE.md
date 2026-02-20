# CLAUDE.md — Agent Instructions for Hamlet

## Project Overview

Hamlet is a multi-framework test converter — 25 conversion directions across 16 frameworks in JavaScript, Java, and Python. It is a Node.js CLI tool and library published to npm. The codebase is JavaScript ES modules with TypeScript type definitions.

## Quick Reference

- **Language:** JavaScript (ES modules — `"type": "module"`)
- **Node:** >= 22.0.0 (CI tests on 22.x and 24.x)
- **Package manager:** npm (`package-lock.json`)
- **Test runner:** Jest with `NODE_OPTIONS='--experimental-vm-modules'`
- **Linter:** ESLint (`eslint:recommended` + `prettier`)
- **Formatter:** Prettier (`.prettierrc`: single quotes, ES5 trailing commas)
- **Commit style:** Conventional Commits (enforced by commitlint via `.husky/commit-msg`)

## Commands

```bash
npm test                    # Run all tests
npm run lint                # Lint source files (src/**/*.js)
npm run format              # Format source files with Prettier (src + bin)
npm run format:check        # Check formatting without writing (used in CI)
npm run test:staged         # Run tests related to staged files
node bin/hamlet.js          # Run the CLI
```

> **Note:** There is no `build` script — the project ships raw ES modules. There is no `validate` script.

## CI Gates

CI (`.github/workflows/ci.yml`) runs these checks in order on every push and PR to `main`/`develop`. **All three are hard gates — any failure blocks the build.**

1. `npm run format:check` — Prettier formatting
2. `npm run lint` — ESLint
3. `npm test` — Jest test suite

### CI-only test skip

`jest.config.js` conditionally skips `test/index.test.js` when `process.env.CI` is set, due to a `cjs-module-lexer` bug with `signal-exit` on Node 18–20. The test runs and passes locally.

### ESM worker-shutdown warning

Jest 29 requires `--experimental-vm-modules` for ESM support. This flag leaves internal IPC Socket handles in every Jest worker, which can cause a "worker process has failed to exit gracefully" warning on suite shutdown. The warning is **not** indicative of a test-level leak — it is a systemic Jest 29 limitation that affects whichever worker happens to shut down last. See `docs/adr/004-jest-esm-strategy.md` for analysis and options.

## Architecture

```
src/
├── cli/               # shorthands.js — shorthand command definitions
├── core/              # 25 modules: BaseConverter, ConverterFactory, PatternEngine,
│                      #   FrameworkDetector, FrameworkRegistry, ConfigConverter,
│                      #   MigrationEngine, Scanner, FileClassifier, ir.js, etc.
├── converters/        # 6 E2E converter classes (Cypress/Playwright/Selenium pairs)
├── converter/         # Batch processing, orchestration, validation, TypeScript support
├── languages/         # Framework definitions organized by language:
│   ├── java/          #   junit4, junit5, testng
│   ├── javascript/    #   cypress, jest, mocha, jasmine, playwright, vitest,
│   │                  #   puppeteer, testcafe, webdriverio
│   └── python/        #   pytest, unittest_fw, nose2
├── patterns/commands/ # Regex pattern definitions (assertions, navigation, selectors, etc.)
├── utils/             # helpers.js (fileUtils, stringUtils, codeUtils, etc.), reporter.js
├── types/             # TypeScript type definitions (index.d.ts)
└── index.js           # Main entry point with public API functions
```

### Key patterns

- **Inheritance:** All converters extend `BaseConverter`. Override `convert()`, `convertConfig()`, `getImports()`, `detectTestTypes()`.
- **Factory:** `ConverterFactory.createConverter(from, to, options)` is async — uses dynamic `import()` to lazy-load converter modules and avoid circular dependencies.
- **PatternEngine:** Registry of regex patterns organized by category with priority-based application. Each converter creates its own engine instance.
- **FrameworkDetector:** Purely static utility class — no instantiation needed.
- **Barrel exports:** `src/core/index.js` and `src/converters/index.js` re-export their modules.

---

## Code Style (Enforced by Prettier + ESLint)

Prettier owns all formatting. ESLint catches logical errors only.

| Concern | Tool | Config |
|---------|------|--------|
| Indentation (2 spaces) | Prettier | default |
| Quotes (single) | Prettier | `.prettierrc` |
| Semicolons (always) | Prettier | default |
| Trailing commas (ES5) | Prettier | `.prettierrc` |
| Print width (80) | Prettier | default |
| Unused variables | ESLint | Error, unless prefixed with `_` |
| `console.log` | ESLint | Allowed (`no-console: off`) |

ESLint extends `eslint:recommended` and `prettier` (via `eslint-config-prettier`, which disables all formatting rules that conflict with Prettier). Do not add formatting rules back into `.eslintrc.json`.

---

## How to Make Changes Safely

1. **Read before writing.** Understand existing code before modifying it.
2. **Small diffs.** One concern per commit. Do not bundle unrelated changes.
3. **Run checks before committing:** `npm run format:check && npm run lint && npm test`
4. **No surprise refactors.** Do not rename, reorganize, or "improve" code outside the scope of the task.
5. **Feature branches.** Never commit directly to `main`. Create a branch, open a PR.

---

## CRITICAL: Rules to Prevent Common LLM Errors

### 1. NEVER use mocks, spies, or jest.mock()

This project tests real implementations exclusively. There are zero `jest.mock()` calls in the codebase. Do not introduce them.

**Why:** Mocks hide real bugs and create maintenance burden. The converters are pure transformations — test the actual output.

```javascript
// WRONG — do not do this
jest.mock('../../src/core/PatternEngine.js');
const mockConvert = jest.fn().mockReturnValue('mocked output');

// CORRECT — use real instances
import { CypressToPlaywright } from '../../src/converters/CypressToPlaywright.js';
const converter = new CypressToPlaywright();
const result = await converter.convert(input);
expect(result).toContain('expected output');
```

The only mock-adjacent pattern is `jest.clearAllMocks()` in the global `test/setup.js` — that is a cleanup safety net, not a testing strategy.

### 2. ALWAYS include `.js` file extensions in imports

This is an ES modules project. All relative imports MUST include the `.js` extension. Omitting it will cause runtime errors.

```javascript
// WRONG
import { BaseConverter } from '../core/BaseConverter';
import { PatternEngine } from '../core/PatternEngine';

// CORRECT
import { BaseConverter } from '../core/BaseConverter.js';
import { PatternEngine } from '../core/PatternEngine.js';
```

### 3. NEVER use nested or circular imports

Import from the direct file path, not through barrel re-exports when inside the same package scope. Do not create import cycles.

```javascript
// WRONG — importing from barrel inside the same package can cause circular deps
import { CypressToPlaywright } from '../converters/index.js';  // from within src/core/

// CORRECT — import directly from the file
import { CypressToPlaywright } from '../converters/CypressToPlaywright.js';
```

The ConverterFactory specifically uses dynamic `import()` to break circular dependencies — do not change this to static imports.

### 4. NEVER use `require()` or CommonJS syntax

The project is `"type": "module"`. All files must use `import`/`export`. No `module.exports`, no `require()`.

```javascript
// WRONG
const { PatternEngine } = require('../core/PatternEngine.js');
module.exports = { MyConverter };

// CORRECT
import { PatternEngine } from '../core/PatternEngine.js';
export class MyConverter extends BaseConverter { }
```

**Exceptions:**
- `commitlint.config.cjs` uses `module.exports` because commitlint requires it.
- `src/index.js` and `bin/hamlet.js` use `createRequire` to read `package.json` (the only supported way to import JSON in ESM without import assertions).
- Converter output strings may contain `require()` / `module.exports` when generating CommonJS target code (e.g., Cypress configs). These are string literals, not module-system usage.

### 5. Use `import` for Node.js builtins — with correct module paths

```javascript
// CORRECT
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import { execFileSync } from 'child_process';

// WRONG
import fs from 'fs';  // Use 'fs/promises' for async operations
```

### 6. Tests MUST follow existing structure

Every test file must follow this exact pattern:

```javascript
import { SomeClass } from '../../src/path/to/SomeClass.js';

describe('SomeClass', () => {
  let instance;

  beforeEach(() => {
    instance = new SomeClass();
  });

  describe('methodName', () => {
    it('should do the expected thing', () => {
      const result = instance.methodName(input);
      expect(result).toBe(expectedOutput);
    });

    it('should handle edge case', () => {
      // ...
    });
  });
});
```

**Rules:**
- Test file naming: `ClassName.test.js` in the matching `test/` subdirectory
- Nest describes: outer = class name, inner = method/feature name
- Use `beforeEach` to create fresh instances — never share mutable state across tests
- Use Jest `expect()` assertions only — no chai, no assert
- Async tests use `async/await`, not callbacks or `.then()`

### 7. Every new public function or class MUST have test coverage

When adding a new converter, utility function, or public method:
- Create a corresponding test file in the matching `test/` subdirectory
- Cover the happy path, at least one edge case, and error conditions
- Coverage thresholds are enforced at 50% for branches, functions, lines, and statements

### 8. NEVER create unnecessary files

Do not create:
- Separate mock files or `__mocks__/` directories
- Test utility/helper modules (inline helpers in the test file if needed)
- New barrel exports without necessity
- `.d.ts` files unless updating the public API surface

### 9. Follow the existing class hierarchy

When adding a new converter:

```javascript
import { BaseConverter } from '../core/BaseConverter.js';
import { PatternEngine } from '../core/PatternEngine.js';

export class NewConverter extends BaseConverter {
  constructor(options = {}) {
    super(options);
    this.sourceFramework = 'source';
    this.targetFramework = 'target';
    this.engine = new PatternEngine();
    this.initializePatterns();
  }

  initializePatterns() {
    // Register patterns with this.engine
  }

  async convert(content, options = {}) {
    // Apply patterns via this.engine.applyPatterns()
    // Return converted string
  }

  async convertConfig(configPath, options = {}) { /* ... */ }
  getImports(testTypes) { /* ... */ }
  detectTestTypes(content) { /* ... */ }
}
```

Then register it in `ConverterFactory` with a lazy `import()` loader.

### 10. Error handling conventions

- Throw `new Error('descriptive message')` — do not use custom error classes
- Include context in error messages (framework names, file paths, what failed)
- Use `addWarning()` and `addError()` on the converter stats for non-fatal issues
- Wrap file I/O in try/catch and re-throw with context

```javascript
// CORRECT
throw new Error(`Failed to convert ${sourcePath}: ${error.message}`);

// WRONG — too vague
throw new Error('Conversion failed');

// WRONG — custom error class (not used in this project)
throw new ConversionError('Failed');
```

### 11. Do not modify configuration files without explicit instruction

Do not change: `jest.config.js`, `.eslintrc.json`, `.prettierrc`, `tsconfig.json`, `playwright.config.js`, `commitlint.config.cjs`, `package.json` scripts, or `.husky/` hooks — unless the task specifically requires it.

### 12. Naming conventions

| Type | Convention | Example |
|------|-----------|---------|
| Classes | PascalCase | `CypressToPlaywright` |
| Methods/functions | camelCase | `detectTestTypes` |
| Constants | UPPER_SNAKE_CASE | `FRAMEWORKS`, `SUPPORTED_TEST_TYPES` |
| Files (classes) | PascalCase matching class | `CypressToPlaywright.js` |
| Files (utilities) | camelCase | `helpers.js`, `batchProcessor.js` |
| Test files | `ClassName.test.js` | `PatternEngine.test.js` |
| Unused params | Prefix with `_` | `(_req, res)` |

### 13. Commit messages

Commits are enforced by commitlint (`.husky/commit-msg`) with conventional commit format:

```
type(scope): description

# Allowed types:
# build, chore, ci, docs, feat, fix, perf, refactor, revert, style, test
```

Examples:
```
feat(converters): add Cypress to Selenium config conversion
fix(pattern-engine): handle regex special characters in selectors
test(factory): add edge case tests for unsupported frameworks
```

> **Note:** Only the `commit-msg` hook is installed. There is no `pre-commit` hook, so `lint-staged` (configured in `package.json`) does not run automatically. Run `npm run format:check && npm run lint` manually before committing.

### 14. Do not introduce new dependencies without justification

Do not add npm packages for functionality that can be implemented in a few lines. Especially avoid:
- Testing utilities (no `@testing-library/*`, no `sinon`, no `nock`)
- Type-checking libraries (the project uses JSDoc + `.d.ts` files)

> **Note:** `lodash` is listed as a runtime dependency in `package.json` but is not currently imported by any source file. Do not add new lodash imports — use native equivalents.

### 15. PR readability checklist

Before submitting any PR, verify:
- [ ] `npm run format:check` passes (zero formatting issues)
- [ ] `npm run lint` passes (zero errors)
- [ ] `npm test` passes (all 705 suites, 1916 tests)
- [ ] No `jest.mock()`, `jest.spyOn()`, or mock files were introduced
- [ ] All imports include `.js` extensions
- [ ] No `require()` or `module.exports` (except `commitlint.config.cjs` and `createRequire` for JSON)
- [ ] New public functions/classes have corresponding tests
- [ ] No unnecessary files were created
- [ ] Commit messages follow conventional commit format
- [ ] No secrets, credentials, or `.env` files are included
- [ ] No `console.log` used for debugging (only intentional chalk-formatted output)

---

## File-by-File Reference

| File | Purpose | Key exports |
|------|---------|-------------|
| `src/index.js` | Main API entry point | `convertFile`, `convertRepository`, `VERSION`, etc. |
| `src/core/BaseConverter.js` | Abstract base class | `BaseConverter` |
| `src/core/ConverterFactory.js` | Factory + lazy loading | `ConverterFactory`, `FRAMEWORKS` |
| `src/core/PatternEngine.js` | Regex pattern registry | `PatternEngine` |
| `src/core/FrameworkDetector.js` | Auto-detect framework | `FrameworkDetector` |
| `src/core/FrameworkRegistry.js` | Framework metadata registry | `FrameworkRegistry` |
| `src/core/ConfigConverter.js` | Config file conversion | `ConfigConverter` |
| `src/core/MigrationEngine.js` | Full project migration | `MigrationEngine` |
| `src/core/Scanner.js` | File discovery | `Scanner` |
| `src/core/FileClassifier.js` | Classify file types | `FileClassifier` |
| `src/core/ir.js` | Intermediate representation | IR node constructors |
| `src/converters/*.js` | 6 E2E converter classes | One class each |
| `src/languages/*/frameworks/*.js` | Framework pattern definitions | Pattern registrations |
| `src/cli/shorthands.js` | CLI shorthand definitions | `SHORTHANDS`, `CONVERSION_CATEGORIES` |
| `src/utils/helpers.js` | Utility namespaces | `fileUtils`, `stringUtils`, `codeUtils`, `testUtils`, `reportUtils`, `logUtils` |
| `config/index.js` | Configuration defaults | Default configs for conversion, reporting, TypeScript |
| `bin/hamlet.js` | CLI entry point | Commander.js CLI |

## Running Tests

```bash
# Full suite
npm test

# Single test file
NODE_OPTIONS='--experimental-vm-modules' npx jest test/core/PatternEngine.test.js

# With coverage
NODE_OPTIONS='--experimental-vm-modules' npx jest --coverage

# Only tests related to changed files
npm run test:staged
```

The `NODE_OPTIONS='--experimental-vm-modules'` flag is required because Jest does not natively support ES modules.
