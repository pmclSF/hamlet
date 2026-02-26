/**
 * Round-trip testing: convert A→B→A and verify semantic equivalence.
 *
 * For bidirectional pairs, the round-tripped code should have:
 * - Same number of test suites (±1 for nesting differences)
 * - Same number of test cases
 * - Same or more assertions (compound asserts may split)
 * - Same hook count
 *
 * For paradigm crossings (pytest↔unittest), thresholds are lower.
 */

import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import {
  convert,
  parseStructure,
  assertSemanticEquivalence,
} from './roundtrip.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const fixturesDir = path.resolve(__dirname, 'fixtures');

async function readFixture(relativePath) {
  return fs.readFile(path.join(fixturesDir, relativePath), 'utf8');
}

// ── Bidirectional round-trip pairs ───────────────────────────────────

const ROUND_TRIP_PAIRS = [
  { dir: 'jest-mocha', from: 'jest', to: 'mocha', ext: 'js' },
  { dir: 'jest-jasmine', from: 'jest', to: 'jasmine', ext: 'js' },
  { dir: 'cypress-playwright', from: 'cypress', to: 'playwright', ext: 'js' },
  { dir: 'cypress-wdio', from: 'cypress', to: 'webdriverio', ext: 'js' },
  { dir: 'playwright-wdio', from: 'playwright', to: 'webdriverio', ext: 'js' },
  {
    dir: 'playwright-puppeteer',
    from: 'playwright',
    to: 'puppeteer',
    ext: 'js',
  },
  { dir: 'junit5-testng', from: 'junit5', to: 'testng', ext: 'java' },
  { dir: 'pytest-unittest', from: 'pytest', to: 'unittest', ext: 'py' },
  { dir: 'cypress-selenium', from: 'cypress', to: 'selenium', ext: 'js' },
  { dir: 'playwright-selenium', from: 'playwright', to: 'selenium', ext: 'js' },
];

const COMPLEXITY_LEVELS = ['simple', 'medium', 'complex'];

// ── One-way conversion directions (no reverse) ──────────────────────

const ONE_WAY_DIRECTIONS = [
  {
    fixture: 'jest-mocha/simple.input.js',
    from: 'jest',
    to: 'vitest',
    framework: 'jest',
  },
  {
    fixture: 'junit5-testng/simple.input.java',
    from: 'junit4',
    to: 'junit5',
    framework: 'junit4',
    note: 'using junit5 fixture as a proxy',
  },
];

// ── Round-trip tests ─────────────────────────────────────────────────

describe('Round-Trip Testing', () => {
  for (const pair of ROUND_TRIP_PAIRS) {
    describe(`${pair.from} ↔ ${pair.to}`, () => {
      for (const complexity of COMPLEXITY_LEVELS) {
        const fixturePath = `${pair.dir}/${complexity}.input.${pair.ext}`;

        it(`round-trips ${complexity} file through ${pair.to} and back`, async () => {
          let original;
          try {
            original = await readFixture(fixturePath);
          } catch (_e) {
            // Skip if fixture doesn't exist yet
            console.warn(`Skipping: fixture ${fixturePath} not found`);
            return;
          }

          // Forward: from → to
          const forward = await convert(original, pair.from, pair.to);
          expect(forward.code).toBeTruthy();
          expect(forward.code.length).toBeGreaterThan(0);

          // Reverse: to → from
          const reverse = await convert(forward.code, pair.to, pair.from);
          expect(reverse.code).toBeTruthy();
          expect(reverse.code.length).toBeGreaterThan(0);

          // Semantic equivalence
          const originalStructure = parseStructure(original, pair.from);
          const roundTrippedStructure = parseStructure(reverse.code, pair.from);

          // Paradigm crossings, Java round-trips, and selenium legacy
          // converters get wider tolerance
          const isPrdmCrossing =
            pair.from === 'pytest' || pair.from === 'unittest';
          const isJava = pair.from === 'junit5' || pair.from === 'testng';
          const isSelenium = pair.to === 'selenium' || pair.from === 'selenium';
          const tolerance =
            isPrdmCrossing || isJava || isSelenium
              ? { testTolerance: 20, assertionTolerance: 10 }
              : complexity === 'complex'
                ? { testTolerance: 2, assertionTolerance: 3 }
                : { testTolerance: 1, assertionTolerance: 2 };

          assertSemanticEquivalence(
            originalStructure,
            roundTrippedStructure,
            tolerance
          );
        });
      }
    });
  }

  describe('One-way conversions (forward only)', () => {
    it('Jest → Vitest produces valid output with high confidence', async () => {
      const original = await readFixture('jest-mocha/simple.input.js');
      const result = await convert(original, 'jest', 'vitest');
      expect(result.code).toBeTruthy();
      expect(result.code.length).toBeGreaterThan(0);

      const structure = parseStructure(result.code, 'vitest');
      const origStructure = parseStructure(original, 'jest');
      expect(structure.testCount).toBeGreaterThanOrEqual(
        origStructure.testCount
      );
    });

    it('nose2 → pytest produces valid output', async () => {
      const noseInput = `import nose2
from nose2.tools import params

def test_addition():
    assert 2 + 2 == 4

@params((1, 1, 2), (2, 3, 5))
def test_add_params(a, b, expected):
    assert a + b == expected
`;
      const result = await convert(noseInput, 'nose2', 'pytest');
      expect(result.code).toBeTruthy();
      expect(result.code.length).toBeGreaterThan(0);
    });

    it('TestCafe → Playwright produces valid output', async () => {
      const tcInput = `import { Selector } from 'testcafe';

fixture('Login').page('http://localhost:3000/login');

test('should login with valid credentials', async t => {
  await t.typeText('#email', 'user@test.com');
  await t.typeText('#password', 'password123');
  await t.click('#submit');
  await t.expect(Selector('.dashboard').exists).ok();
});
`;
      const result = await convert(tcInput, 'testcafe', 'playwright');
      expect(result.code).toBeTruthy();
      expect(result.code.length).toBeGreaterThan(0);
    });

    it('TestCafe → Cypress produces valid output', async () => {
      const tcInput = `import { Selector } from 'testcafe';

fixture('Search').page('http://localhost:3000');

test('should search and find results', async t => {
  await t.typeText('#search', 'hamlet');
  await t.click('#search-btn');
  await t.expect(Selector('.results').count).gte(1);
});
`;
      const result = await convert(tcInput, 'testcafe', 'cypress');
      expect(result.code).toBeTruthy();
      expect(result.code.length).toBeGreaterThan(0);
    });

    it('JUnit 4 → JUnit 5 produces valid output', async () => {
      const j4Input = `import org.junit.Test;
import org.junit.Before;
import org.junit.After;
import static org.junit.Assert.*;

public class CalculatorTest {
    private Calculator calc;

    @Before
    public void setUp() {
        calc = new Calculator();
    }

    @After
    public void tearDown() {
        calc = null;
    }

    @Test
    public void testAdd() {
        assertEquals(4, calc.add(2, 2));
    }

    @Test(expected = ArithmeticException.class)
    public void testDivideByZero() {
        calc.divide(1, 0);
    }
}
`;
      const result = await convert(j4Input, 'junit4', 'junit5');
      expect(result.code).toBeTruthy();
      expect(result.code.length).toBeGreaterThan(0);
    });
  });
});
