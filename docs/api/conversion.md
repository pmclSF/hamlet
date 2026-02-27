# Conversion Process

## Overview

Hamlet converts test files between frameworks using a multi-step pipeline:

1. **Detect** - determine the source framework from file content using regex heuristics
2. **Parse** - classify source lines into IR (intermediate representation) nodes: suites, tests, hooks, assertions, raw code
3. **Transform** - apply regex-based pattern substitutions to convert API calls and test structure
4. **Score** - walk the IR tree to calculate confidence (converted vs. unconvertible nodes)
5. **Report** - generate HAMLET-TODO markers for patterns that need manual review

## Pattern Engine

The `PatternEngine` is the core of the conversion pipeline. Each converter creates its own engine instance and registers patterns organized by category:

- **Assertions**: `expect(...).toBe()` &rarr; `assert.equal()`
- **Navigation**: `cy.visit()` &rarr; `page.goto()`
- **Selectors**: `cy.get()` &rarr; `page.locator()`
- **Lifecycle**: `beforeEach` &rarr; `@BeforeEach` (Java)
- **Imports**: Framework-specific import statements

Patterns have priorities that control application order, ensuring dependent transformations happen in the correct sequence.

## Intermediate Representation (IR)

The IR captures test structure as a tree of nodes:

- **Suite** - `describe` / `test.describe` blocks
- **Test** - individual test cases
- **Hook** - `beforeEach`, `afterAll`, etc.
- **Assertion** - `expect(...)` calls
- **Raw** - code lines that pass through unchanged

The IR is used for confidence scoring and structure analysis, not for code emission. The actual output is produced by applying regex patterns to the source string.

## Confidence Scoring

Every conversion produces a confidence score (0-100%):

| Range | Meaning |
|-------|---------|
| 80-100% | Fully automated, ready to use |
| 50-79% | Mostly automated, review HAMLET-TODO markers |
| 0-49% | Significant manual work needed |

The score is calculated by comparing:
- Lines with successful pattern matches (converted)
- Lines flagged as unconvertible (HAMLET-TODO)
- Lines that pass through unchanged (neutral)

## Supported Conversion Examples

### JavaScript Unit: Jest to Vitest

```javascript
// Jest
import { render } from '@testing-library/react';

describe('Button', () => {
  it('should render', () => {
    expect(true).toBe(true);
  });
});

// Vitest (converted)
import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';

describe('Button', () => {
  it('should render', () => {
    expect(true).toBe(true);
  });
});
```

### JavaScript E2E: Cypress to Playwright

```javascript
// Cypress
describe('Login', () => {
  it('should log in', () => {
    cy.visit('/login');
    cy.get('#username').type('admin');
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
    await page.locator('button').click();
    await expect(page).toHaveURL(/dashboard/);
  });
});
```

### Java: JUnit 4 to JUnit 5

```java
// JUnit 4
import org.junit.Test;
import org.junit.Before;
import static org.junit.Assert.*;

public class CalcTest {
    @Before
    public void setUp() { }

    @Test
    public void testAdd() {
        assertEquals(4, Calculator.add(2, 2));
    }
}

// JUnit 5 (converted)
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.BeforeEach;
import static org.junit.jupiter.api.Assertions.*;

class CalcTest {
    @BeforeEach
    void setUp() { }

    @Test
    void testAdd() {
        assertEquals(4, Calculator.add(2, 2));
    }
}
```

### Python: pytest to unittest

```python
# pytest
import pytest

def test_add():
    assert add(2, 2) == 4

@pytest.fixture
def client():
    return Client()

def test_client(client):
    assert client.status == 200

# unittest (converted)
import unittest

class TestModule(unittest.TestCase):
    def setUp(self):
        self.client = Client()

    def test_add(self):
        self.assertEqual(add(2, 2), 4)

    def test_client(self):
        self.assertEqual(self.client.status, 200)
```

## HAMLET-TODO Markers

When a pattern cannot be automatically converted, Hamlet inserts a marker:

```javascript
// HAMLET-TODO: cy.intercept() has no direct equivalent in Vitest
// Original: cy.intercept('POST', '/api/login').as('login')
```

Common reasons for HAMLET-TODO markers:
- Framework-specific APIs with no target equivalent
- Custom commands or plugins
- Complex assertion chains
- Dynamic selectors or programmatic patterns

## Config Conversion

Framework config files can also be converted:

```bash
hamlet convert-config jest.config.js --to vitest -o vitest.config.js
hamlet convert-config cypress.config.js --to playwright -o playwright.config.ts
```

Config conversion handles:
- Test file patterns and paths
- Plugin and preset mappings
- Timeout and retry settings
- Reporter configuration
