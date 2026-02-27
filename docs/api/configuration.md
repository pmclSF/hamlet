# Configuration

Hamlet is configured through CLI flags and programmatic API options. There is no configuration file format.

## CLI Flags

All conversion behavior is controlled via command-line flags:

| Flag | Description | Default |
|------|-------------|---------|
| `-f, --from <framework>` | Source framework | (required, or use `--auto-detect`) |
| `-t, --to <framework>` | Target framework | (required) |
| `-o, --output <path>` | Output path | (required for directories) |
| `--dry-run` | Preview without writing files | `false` |
| `--on-error <mode>` | Error handling mode | `skip` |
| `-q, --quiet` | Suppress non-error output | `false` |
| `--verbose` | Show detailed per-pattern output | `false` |
| `--json` | Machine-readable JSON output | `false` |
| `--no-color` | Disable colored output | `false` |
| `--debug` | Show stack traces on error | `false` |
| `--auto-detect` | Auto-detect source framework | `false` |

### Error handling modes

The `--on-error` flag controls what happens when a file fails to convert:

- **`skip`** (default): Skip the failed file and continue with the rest
- **`fail`**: Stop immediately on the first error
- **`best-effort`**: Write partial output even for files that error

### Environment variables

| Variable | Description |
|----------|-------------|
| `NO_COLOR` | Disable colored output (equivalent to `--no-color`) |
| `DEBUG` | Enable debug output (equivalent to `--debug`) |

## Programmatic API

### ConverterFactory

Create converters programmatically using `ConverterFactory`:

```javascript
import { ConverterFactory } from 'hamlet-converter/core';

const converter = await ConverterFactory.createConverter('jest', 'vitest');
const output = await converter.convert(sourceCode);
```

`createConverter(from, to, options)` accepts:
- `from` (string): Source framework name
- `to` (string): Target framework name
- `options` (object, optional): Converter options

### Converter options

Options passed to `createConverter` or directly to a converter constructor:

```javascript
const converter = await ConverterFactory.createConverter('jest', 'vitest', {
  verbose: false,
  preserveStructure: true,
});
```

### convertFile

Convert a single file using the high-level API:

```javascript
import { convertFile } from 'hamlet-converter';

const result = await convertFile('auth.test.js', {
  from: 'jest',
  to: 'vitest',
  outputPath: 'converted/auth.test.js',
});
```

### convertRepository

Convert an entire directory:

```javascript
import { convertRepository } from 'hamlet-converter';

const results = await convertRepository('tests/', {
  from: 'jest',
  to: 'vitest',
  outputDir: 'converted/',
});
```

### Conversion report

After conversion, get a report with confidence score and statistics:

```javascript
const converter = await ConverterFactory.createConverter('jest', 'vitest');
const output = await converter.convert(sourceCode);
const report = converter.getLastReport();

console.log(`Confidence: ${report.confidence}%`);
console.log(`Patterns applied: ${report.patternsApplied}`);
console.log(`Warnings: ${report.warnings.length}`);
```

## Supported Frameworks

Use `FRAMEWORKS` to get the list of supported framework identifiers:

```javascript
import { FRAMEWORKS } from 'hamlet-converter/core';
console.log(FRAMEWORKS);
// ['cypress', 'playwright', 'selenium', 'jest', 'vitest', 'mocha', 'jasmine',
//  'junit4', 'junit5', 'testng', 'pytest', 'unittest', 'nose2',
//  'webdriverio', 'puppeteer', 'testcafe']
```
