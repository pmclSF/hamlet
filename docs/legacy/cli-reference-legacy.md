> **Legacy document.** This describes the legacy JavaScript converter engine. For the current engine, see the [CLI spec](../cli-spec.md) and [architecture overview](../architecture/00-overview.md).

# CLI Reference (Legacy Converter)

## Commands

### `convert`

Convert test files from one framework to another.

```bash
terrain convert <source> [options]
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
terrain convert auth.test.js --from jest --to vitest -o converted/

# Directory
terrain convert tests/ --from jest --to vitest -o converted/

# Glob pattern
terrain convert "tests/**/*.test.js" --from jest --to vitest -o converted/

# Dry run
terrain convert tests/ --from jest --to vitest -o converted/ --dry-run
```

### Shorthands

Shorthands are aliases for `convert --from X --to Y`:

```bash
# These are equivalent:
terrain convert auth.test.js --from jest --to vitest -o converted/
terrain jest2vt auth.test.js -o converted/
```

Run `terrain shorthands` to list all shorthand aliases.

### `migrate`

Full project migration with state tracking, dependency ordering, and config conversion.

```bash
terrain migrate <source> [options]
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
terrain migrate tests/ --from jest --to vitest -o converted/

# Resume interrupted migration
terrain migrate tests/ --from jest --to vitest -o converted/ --continue

# Retry failed files only
terrain migrate tests/ --from jest --to vitest -o converted/ --retry-failed
```

### `estimate`

Preview migration complexity without converting.

```bash
terrain estimate <source> --from <framework> --to <framework>
```

**Example:**
```bash
terrain estimate tests/ --from jest --to vitest
```

### `convert-config`

Convert framework configuration files.

```bash
terrain convert-config <config-file> --to <framework> -o <output>
```

**Example:**
```bash
terrain convert-config jest.config.js --to vitest -o vitest.config.js
terrain convert-config cypress.config.js --to playwright -o playwright.config.ts
```

### `list`

Show all supported conversion directions with their shorthand aliases.

```bash
terrain list
```

### `shorthands`

List all shorthand command aliases.

```bash
terrain shorthands
```

### `detect`

Auto-detect the testing framework used in a file.

```bash
terrain detect <file>
```

**Example:**
```bash
terrain detect auth.test.js
# Output: jest (confidence: 95%)
```

### `doctor`

Run diagnostics to verify your environment.

```bash
terrain doctor
```

### `status`

Show current migration progress.

```bash
terrain status -d <directory>
```

### `checklist`

Generate a migration checklist.

```bash
terrain checklist -d <directory>
```

### `reset`

Clear migration state.

```bash
terrain reset -d <directory> --yes
```

### `serve`

Start the API server. Exposes REST endpoints for programmatic access but does not serve the browser UI.

```bash
terrain serve [options]
```

**Options:**
- `-p, --port <number>`: Port to listen on (0 = random, default: 0)
- `--root <path>`: Project root directory (default: `.`)

**Example:**
```bash
terrain serve --port 3000 --root ./my-project
```

### `ui`

Start the server with the browser UI and open it automatically.

```bash
terrain ui [options]
```

**Options:**
- `-p, --port <number>`: Port to listen on (0 = random, default: 0)
- `--root <path>`: Project root directory (default: `.`)
- `--no-open`: Don't auto-open browser

**Example:**
```bash
terrain ui --root ./my-project
terrain ui --port 8080 --no-open
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
terrain jest2vt auth.test.js -o converted/ --json
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
