# CLI Documentation

## Commands

### `convert`
Convert Cypress tests to Playwright format.

```bash
hamlet convert <source> [options]
```

#### Arguments
- `source`: Path to Cypress test file or directory

#### Options
- `-o, --output <path>`: Output path for converted tests
- `-c, --config <path>`: Custom configuration file path
- `-t, --test-type <type>`: Specify test type (e2e, component, api, etc.)
- `--validate`: Validate converted tests
- `--report <format>`: Generate conversion report (html, json, markdown)
- `--preserve-structure`: Maintain original directory structure
- `--batch-size <number>`: Number of tests to process in parallel
- `--ignore <pattern>`: Files to ignore (glob pattern)
- `--hooks`: Convert test hooks and configurations
- `--plugins`: Convert Cypress plugins

### Examples

```bash
# Convert a single test file
hamlet convert cypress/e2e/login.cy.js

# Convert all tests in a directory
hamlet convert cypress/e2e --output playwright/tests

# Convert and validate
hamlet convert cypress/e2e --validate --report html

# Convert repository from GitHub
hamlet convert https://github.com/user/repo.git
```

### Error Handling

The CLI will provide detailed error messages in case of issues:
```bash
✗ Error converting file: syntax error in login.cy.js
  - Line 15: Unexpected token
  - Suggestion: Check for missing brackets or semicolons
```