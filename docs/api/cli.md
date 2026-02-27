# CLI Reference

## Commands

### `convert`

Convert test files from one framework to another.

```bash
hamlet convert <source> [options]
```

**Arguments:**
- `source`: Path to a test file, directory, or glob pattern

**Options:**
- `-f, --from <framework>`: Source framework
- `-t, --to <framework>`: Target framework
- `-o, --output <path>`: Output path (required for directories)
- `--dry-run`: Preview without writing files
- `--on-error <mode>`: Error handling: `skip` (default), `fail`, `best-effort`
- `--auto-detect`: Auto-detect source framework

**Examples:**
```bash
# Single file
hamlet convert auth.test.js --from jest --to vitest -o converted/

# Directory
hamlet convert tests/ --from jest --to vitest -o converted/

# Glob pattern
hamlet convert "tests/**/*.test.js" --from jest --to vitest -o converted/

# Dry run
hamlet convert tests/ --from jest --to vitest -o converted/ --dry-run
```

### Shorthands

Shorthands are aliases for `convert --from X --to Y`:

```bash
# These are equivalent:
hamlet convert auth.test.js --from jest --to vitest -o converted/
hamlet jest2vt auth.test.js -o converted/
```

Run `hamlet shorthands` to list all shorthand aliases.

### `migrate`

Full project migration with state tracking, dependency ordering, and config conversion.

```bash
hamlet migrate <source> [options]
```

**Options:**
- `-f, --from <framework>`: Source framework
- `-t, --to <framework>`: Target framework
- `-o, --output <path>`: Output directory
- `--dry-run`: Preview without writing files
- `--continue`: Resume an interrupted migration
- `--retry-failed`: Retry only previously failed files

**Examples:**
```bash
# Migrate a project
hamlet migrate tests/ --from jest --to vitest -o converted/

# Resume interrupted migration
hamlet migrate tests/ --from jest --to vitest -o converted/ --continue

# Retry failed files only
hamlet migrate tests/ --from jest --to vitest -o converted/ --retry-failed
```

### `estimate`

Preview migration complexity without converting.

```bash
hamlet estimate <source> --from <framework> --to <framework>
```

**Example:**
```bash
hamlet estimate tests/ --from jest --to vitest
```

### `convert-config`

Convert framework configuration files.

```bash
hamlet convert-config <config-file> --to <framework> -o <output>
```

**Example:**
```bash
hamlet convert-config jest.config.js --to vitest -o vitest.config.js
hamlet convert-config cypress.config.js --to playwright -o playwright.config.ts
```

### `list`

Show all supported conversion directions with their shorthand aliases.

```bash
hamlet list
```

### `shorthands`

List all shorthand command aliases.

```bash
hamlet shorthands
```

### `detect`

Auto-detect the testing framework used in a file.

```bash
hamlet detect <file>
```

**Example:**
```bash
hamlet detect auth.test.js
# Output: jest (confidence: 95%)
```

### `doctor`

Run diagnostics to verify your environment.

```bash
hamlet doctor
```

### `status`

Show current migration progress.

```bash
hamlet status -d <directory>
```

### `checklist`

Generate a migration checklist.

```bash
hamlet checklist -d <directory>
```

### `reset`

Clear migration state.

```bash
hamlet reset -d <directory> --yes
```

### `serve`

Start the built-in web UI server for interactive conversion.

```bash
hamlet serve [options]
```

### `open`

Start the server and open the web UI in your browser.

```bash
hamlet open [options]
```

## Global Options

| Flag | Description |
|------|-------------|
| `-o, --output <path>` | Output path (required for directories) |
| `-f, --from <framework>` | Source framework |
| `-t, --to <framework>` | Target framework |
| `--dry-run` | Preview without writing files |
| `--on-error <mode>` | Error handling: `skip` (default), `fail`, `best-effort` |
| `-q, --quiet` | Suppress non-error output |
| `--verbose` | Show detailed per-pattern output |
| `--json` | Machine-readable JSON output |
| `--no-color` | Disable colored output (also respects `NO_COLOR` env) |
| `--debug` | Show stack traces on error (also respects `DEBUG=1` env) |
| `--auto-detect` | Auto-detect source framework |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error (conversion failed, unexpected exception) |
| 2 | Invalid arguments (bad framework, missing file, missing flags) |

## JSON Output

For CI integration, use `--json` for machine-readable output:

```bash
hamlet jest2vt auth.test.js -o converted/ --json
```

```json
{
  "success": true,
  "files": [
    {
      "source": "auth.test.js",
      "output": "converted/auth.test.js",
      "confidence": 95
    }
  ],
  "summary": {
    "converted": 1,
    "skipped": 0,
    "failed": 0
  }
}
```

## Error Output

All errors are prefixed with `Error:` and written to stderr. Common errors include a
`Next steps:` hint. Use `--debug` to see full stack traces.
