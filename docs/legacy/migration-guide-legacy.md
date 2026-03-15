> **Legacy document.** This describes the legacy JavaScript converter engine. For the current engine, see the [CLI spec](../cli-spec.md) and [architecture overview](../architecture/00-overview.md).

# Migration Guide (Legacy Converter)

## Overview

This guide walks through migrating an entire test suite from one framework to another using `terrain migrate`. While `terrain convert` handles individual files, `terrain migrate` adds state tracking, dependency ordering, and config conversion for full project migrations.

## Migration Steps

### 1. Estimate the migration

Before converting anything, preview the scope:

```bash
terrain estimate tests/ --from jest --to vitest
```

This reports how many files would be converted, estimated confidence, and any patterns that will need manual attention.

### 2. Run the migration

```bash
terrain migrate tests/ --from jest --to vitest -o converted/
```

The `migrate` command:
- Discovers all test files in the source directory
- Converts each file with pattern-based transformation
- Tracks state so you can resume if interrupted
- Reports per-file confidence scores

### 3. Review TERRAIN-TODO markers

Search converted files for patterns that need manual review:

```bash
grep -r "TERRAIN-TODO" converted/
```

Each marker includes the original code and a description of what needs attention:

```javascript
// TERRAIN-TODO: Custom command has no direct equivalent
// Original: cy.login('admin')
```

### 4. Check migration status

```bash
terrain status -d .
```

This shows how many files have been converted, how many remain, and overall progress.

### 5. Generate a checklist

```bash
terrain checklist -d .
```

Produces a checklist of remaining manual steps for the migration.

## Resuming an Interrupted Migration

If the migration is interrupted, resume from where you left off:

```bash
terrain migrate tests/ --from jest --to vitest -o converted/ --continue
```

To retry only previously failed files:

```bash
terrain migrate tests/ --from jest --to vitest -o converted/ --retry-failed
```

## Resetting Migration State

To clear migration state and start over:

```bash
terrain reset -d . --yes
```

## Dry Run

Preview what a migration would do without writing any files:

```bash
terrain migrate tests/ --from jest --to vitest --dry-run
```

## Common Migration Patterns

### JavaScript unit tests (Jest to Vitest)

```javascript
// Jest
import { render } from '@testing-library/react';

describe('Button', () => {
  it('should render', () => {
    const { getByText } = render(<Button />);
    expect(getByText('Click')).toBeInTheDocument();
  });
});

// Vitest (converted)
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';

describe('Button', () => {
  it('should render', () => {
    const { getByText } = render(<Button />);
    expect(getByText('Click')).toBeInTheDocument();
  });
});
```

### E2E tests (Cypress to Playwright)

```javascript
// Cypress
describe('Login', () => {
  it('should log in', () => {
    cy.visit('/login');
    cy.get('#username').type('admin');
    cy.get('#password').type('secret');
    cy.get('button').click();
    cy.url().should('include', '/dashboard');
  });
});

// Playwright (converted)
import { test, expect } from '@playwright/test';

test.describe('Login', () => {
  test('should log in', async ({ page }) => {
    await page.goto('/login');
    await page.locator('#username').fill('admin');
    await page.locator('#password').fill('secret');
    await page.locator('button').click();
    await expect(page).toHaveURL(/dashboard/);
  });
});
```

### Java (JUnit 4 to JUnit 5)

```java
// JUnit 4
import org.junit.Test;
import static org.junit.Assert.assertEquals;

public class CalcTest {
    @Test
    public void testAdd() {
        assertEquals(4, Calculator.add(2, 2));
    }
}

// JUnit 5 (converted)
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertEquals;

class CalcTest {
    @Test
    void testAdd() {
        assertEquals(4, Calculator.add(2, 2));
    }
}
```

### Python (pytest to unittest)

```python
# pytest
def test_add():
    assert add(2, 2) == 4

def test_subtract():
    assert subtract(5, 3) == 2

# unittest (converted)
import unittest

class TestMath(unittest.TestCase):
    def test_add(self):
        self.assertEqual(add(2, 2), 4)

    def test_subtract(self):
        self.assertEqual(subtract(5, 3), 2)
```

## Confidence Scores

Every conversion produces a confidence score (0-100%):

- **High (80-100%)**: Fully automated, ready to use
- **Medium (50-79%)**: Mostly automated, review TERRAIN-TODO markers
- **Low (0-49%)**: Significant manual work needed

## Troubleshooting

### TERRAIN-TODO markers remain after conversion

This is expected. TERRAIN-TODO markers flag patterns that cannot be automatically converted and need manual attention. Common examples:

- Custom commands / plugins with no target equivalent
- Framework-specific APIs (e.g., `cy.intercept()`, `cy.session()`)
- Complex assertion chains

### Low confidence score

A low score means many patterns in the source file couldn't be mapped to the target framework. Check for:

- Heavy use of framework-specific plugins
- Custom helper abstractions wrapping framework APIs
- Non-standard test patterns

### Conversion produces empty or minimal output

Verify the source and target frameworks are correct:

```bash
terrain detect myfile.test.js
```

If the detector misidentifies the framework, specify it explicitly with `--from`.

## Next Steps

- [CLI Reference](./cli-reference-legacy.md) - all commands and options
- [Configuration](./configuration-legacy.md) - flags and programmatic API
- [Conversion Process](./conversion-process-legacy.md) - how conversion works
