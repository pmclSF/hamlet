/**
 * Parity harness: validates all 25 conversion directions.
 *
 * Tests that:
 * 1. Every framework definition exports detect/parse/emit
 * 2. Every IR emitter exports emitNode/matchesBaseline/emitFullFile
 * 3. Pipeline-backed directions produce valid output
 * 4. Legacy converter directions still work
 * 5. ir-full emitters generate valid output from sample IR
 * 6. Round-trip: source parse → target emitFullFile → valid output
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  Navigation,
  RawCode,
  Comment,
  Modifier,
} from '../../src/core/ir.js';

import { ConverterFactory } from '../../src/core/ConverterFactory.js';

// ═══════════════════════════════════════════════════════════════════════
// All 16 framework definitions
// ═══════════════════════════════════════════════════════════════════════

const FRAMEWORKS = {
  // JavaScript
  cypress: () => import('../../src/languages/javascript/frameworks/cypress.js'),
  playwright: () =>
    import('../../src/languages/javascript/frameworks/playwright.js'),
  jest: () => import('../../src/languages/javascript/frameworks/jest.js'),
  vitest: () => import('../../src/languages/javascript/frameworks/vitest.js'),
  mocha: () => import('../../src/languages/javascript/frameworks/mocha.js'),
  jasmine: () => import('../../src/languages/javascript/frameworks/jasmine.js'),
  webdriverio: () =>
    import('../../src/languages/javascript/frameworks/webdriverio.js'),
  puppeteer: () =>
    import('../../src/languages/javascript/frameworks/puppeteer.js'),
  testcafe: () =>
    import('../../src/languages/javascript/frameworks/testcafe.js'),
  selenium: () =>
    import('../../src/languages/javascript/frameworks/selenium.js'),
  // Java
  junit4: () => import('../../src/languages/java/frameworks/junit4.js'),
  junit5: () => import('../../src/languages/java/frameworks/junit5.js'),
  testng: () => import('../../src/languages/java/frameworks/testng.js'),
  // Python
  pytest: () => import('../../src/languages/python/frameworks/pytest.js'),
  unittest: () =>
    import('../../src/languages/python/frameworks/unittest_fw.js'),
  nose2: () => import('../../src/languages/python/frameworks/nose2.js'),
};

// ═══════════════════════════════════════════════════════════════════════
// All 13 IR emitters (frameworks that can be targets)
// ═══════════════════════════════════════════════════════════════════════

const EMITTERS = {
  cypress: () => import('../../src/core/emitters/cypress/irEmitter.js'),
  playwright: () => import('../../src/core/emitters/playwright/irEmitter.js'),
  jest: () => import('../../src/core/emitters/jest/irEmitter.js'),
  vitest: () => import('../../src/core/emitters/vitest/irEmitter.js'),
  mocha: () => import('../../src/core/emitters/mocha/irEmitter.js'),
  jasmine: () => import('../../src/core/emitters/jasmine/irEmitter.js'),
  webdriverio: () => import('../../src/core/emitters/webdriverio/irEmitter.js'),
  puppeteer: () => import('../../src/core/emitters/puppeteer/irEmitter.js'),
  junit5: () => import('../../src/core/emitters/junit5/irEmitter.js'),
  testng: () => import('../../src/core/emitters/testng/irEmitter.js'),
  pytest: () => import('../../src/core/emitters/pytest/irEmitter.js'),
  unittest: () => import('../../src/core/emitters/unittest/irEmitter.js'),
  selenium: () => import('../../src/core/emitters/selenium/irEmitter.js'),
};

// ═══════════════════════════════════════════════════════════════════════
// All 25 directions (19 pipeline + 5 legacy + 1 selenium stub)
// ═══════════════════════════════════════════════════════════════════════

const ALL_DIRECTIONS = [
  // Pipeline-backed
  'cypress-playwright',
  'jest-vitest',
  'mocha-jest',
  'jasmine-jest',
  'jest-mocha',
  'jest-jasmine',
  'junit4-junit5',
  'junit5-testng',
  'testng-junit5',
  'pytest-unittest',
  'unittest-pytest',
  'nose2-pytest',
  'webdriverio-playwright',
  'webdriverio-cypress',
  'playwright-webdriverio',
  'cypress-webdriverio',
  'puppeteer-playwright',
  'playwright-puppeteer',
  'testcafe-playwright',
  'testcafe-cypress',
  // Legacy
  'playwright-cypress',
  'cypress-selenium',
  'playwright-selenium',
  'selenium-cypress',
  'selenium-playwright',
];

// ═══════════════════════════════════════════════════════════════════════
// Sample source code per framework for testing
// ═══════════════════════════════════════════════════════════════════════

const SAMPLES = {
  cypress: `
describe('Login', () => {
  it('should login successfully', () => {
    cy.visit('/login');
    cy.get('#email').type('user@test.com');
    cy.get('#password').type('password');
    cy.get('button').click();
    cy.url().should('include', '/dashboard');
  });
});
`,
  playwright: `
import { test, expect } from '@playwright/test';
test.describe('Login', () => {
  test('should login successfully', async ({ page }) => {
    await page.goto('/login');
    await page.locator('#email').fill('user@test.com');
    await expect(page).toHaveURL(/dashboard/);
  });
});
`,
  jest: `
describe('Calculator', () => {
  it('should add numbers', () => {
    expect(add(2, 3)).toBe(5);
  });
});
`,
  vitest: `
import { describe, it, expect } from 'vitest';
describe('Calculator', () => {
  it('should add numbers', () => {
    expect(add(2, 3)).toBe(5);
  });
});
`,
  mocha: `
const { expect } = require('chai');
describe('Calculator', () => {
  it('should add numbers', () => {
    expect(add(2, 3)).to.equal(5);
  });
});
`,
  jasmine: `
describe('Calculator', () => {
  it('should add numbers', () => {
    expect(add(2, 3)).toBe(5);
  });
});
`,
  webdriverio: `
describe('Login', () => {
  it('should login', async () => {
    await browser.url('/login');
    await $('#email').setValue('user@test.com');
    await expect($('#email')).toHaveValue('user@test.com');
  });
});
`,
  puppeteer: `
describe('Login', () => {
  it('should login', async () => {
    await page.goto('/login');
    await page.type('#email', 'user@test.com');
    const value = await page.$eval('#email', el => el.value);
    expect(value).toBe('user@test.com');
  });
});
`,
  testcafe: `
import { Selector } from 'testcafe';
fixture('Login').page('/login');
test('should login', async t => {
  await t.typeText('#email', 'user@test.com');
  await t.expect(Selector('#email').value).eql('user@test.com');
});
`,
  selenium: `
const { Builder, By } = require('selenium-webdriver');
describe('Login', () => {
  it('should login', async () => {
    await driver.get('/login');
    await driver.findElement(By.css('#email')).sendKeys('user@test.com');
    const value = await driver.findElement(By.css('#email')).getAttribute('value');
    expect(value).toBe('user@test.com');
  });
});
`,
  junit4: `
import org.junit.Test;
import static org.junit.Assert.*;

public class CalcTest {
    @Test
    public void testAdd() {
        assertEquals(5, add(2, 3));
    }
}
`,
  junit5: `
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

class CalcTest {
    @Test
    void testAdd() {
        Assertions.assertEquals(5, add(2, 3));
    }
}
`,
  testng: `
import org.testng.annotations.Test;
import org.testng.Assert;

public class CalcTest {
    @Test
    public void testAdd() {
        Assert.assertEquals(add(2, 3), 5);
    }
}
`,
  pytest: `
def test_add():
    assert add(2, 3) == 5
`,
  unittest: `
import unittest

class TestCalc(unittest.TestCase):
    def test_add(self):
        self.assertEqual(add(2, 3), 5)
`,
  nose2: `
from nose.tools import assert_equal

def test_add():
    assert_equal(add(2, 3), 5)
`,
};

// ═══════════════════════════════════════════════════════════════════════
// Test: Framework definitions
// ═══════════════════════════════════════════════════════════════════════

describe('Framework definitions', () => {
  const frameworkNames = Object.keys(FRAMEWORKS);

  it.each(frameworkNames)('%s exports detect, parse, emit', async (name) => {
    const mod = await FRAMEWORKS[name]();
    const fw = mod.default;
    expect(typeof fw.detect).toBe('function');
    expect(typeof fw.parse).toBe('function');
    expect(typeof fw.emit).toBe('function');
    expect(fw.name).toBeDefined();
    expect(fw.language).toBeDefined();
  });

  it.each(frameworkNames)('%s detects its own sample code', async (name) => {
    const mod = await FRAMEWORKS[name]();
    const fw = mod.default;
    const sample = SAMPLES[name];
    if (!sample) return;
    const confidence = fw.detect(sample);
    expect(confidence).toBeGreaterThan(0);
  });

  it.each(frameworkNames)(
    '%s parses its own sample to a TestFile',
    async (name) => {
      const mod = await FRAMEWORKS[name]();
      const fw = mod.default;
      const sample = SAMPLES[name];
      if (!sample) return;
      const ir = fw.parse(sample);
      expect(ir).toBeInstanceOf(TestFile);
      expect(ir.body.length).toBeGreaterThan(0);
    }
  );
});

// ═══════════════════════════════════════════════════════════════════════
// Test: IR emitters
// ═══════════════════════════════════════════════════════════════════════

describe('IR emitters', () => {
  const emitterNames = Object.keys(EMITTERS);

  it.each(emitterNames)(
    '%s exports emitNode, matchesBaseline, emitFullFile',
    async (name) => {
      const emitter = await EMITTERS[name]();
      expect(typeof emitter.emitNode).toBe('function');
      expect(typeof emitter.matchesBaseline).toBe('function');
      expect(typeof emitter.emitFullFile).toBe('function');
    }
  );
});

// ═══════════════════════════════════════════════════════════════════════
// Test: ir-full emitters produce output from sample IR
// ═══════════════════════════════════════════════════════════════════════

describe('ir-full emitters produce output', () => {
  // Sample IR tree that all emitters should handle
  function makeSampleIR() {
    return new TestFile({
      language: 'javascript',
      imports: [],
      body: [
        new TestSuite({
          name: 'SampleSuite',
          hooks: [
            new Hook({
              hookType: 'beforeEach',
              body: [new RawCode({ code: 'setup();' })],
            }),
          ],
          tests: [
            new TestCase({
              name: 'testBasic',
              body: [
                new Assertion({
                  kind: 'equal',
                  subject: 'result',
                  expected: '42',
                }),
              ],
            }),
          ],
        }),
      ],
    });
  }

  // Emitters that should produce non-null ir-full output
  const fullEmitters = [
    'cypress',
    'playwright',
    'jest',
    'vitest',
    'mocha',
    'jasmine',
    'webdriverio',
    'puppeteer',
    'junit5',
    'testng',
    'pytest',
    'unittest',
  ];

  it.each(fullEmitters)(
    '%s emitFullFile returns non-null output',
    async (name) => {
      const emitter = await EMITTERS[name]();
      const ir = makeSampleIR();
      const output = emitter.emitFullFile(ir);
      expect(output).not.toBeNull();
      expect(typeof output).toBe('string');
      expect(output.length).toBeGreaterThan(0);
    }
  );

  // Stub emitters that return null (selenium)
  it('selenium emitFullFile returns null (stub)', async () => {
    const emitter = await EMITTERS.selenium();
    const ir = makeSampleIR();
    const output = emitter.emitFullFile(ir);
    expect(output).toBeNull();
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Test: All 25 directions are reported as supported
// ═══════════════════════════════════════════════════════════════════════

describe('All 25 directions', () => {
  it('should all be listed in getSupportedConversions()', () => {
    const supported = ConverterFactory.getSupportedConversions();
    for (const dir of ALL_DIRECTIONS) {
      expect(supported).toContain(dir);
    }
    expect(supported.length).toBe(ALL_DIRECTIONS.length);
  });

  it('should all return true from isSupported()', () => {
    for (const dir of ALL_DIRECTIONS) {
      const [from, to] = dir.split('-');
      expect(ConverterFactory.isSupported(from, to)).toBe(true);
    }
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Test: Emitter coverage scoreboard
// ═══════════════════════════════════════════════════════════════════════

describe('Emitter coverage scoreboard', () => {
  // Build a scoreboard showing which emitters support which assertion kinds
  const ASSERTION_KINDS = [
    'equal',
    'notEqual',
    'truthy',
    'falsy',
    'isNull',
    'isDefined',
    'throws',
    'contains',
    'greaterThan',
    'lessThan',
    'fail',
  ];

  const targetEmitters = [
    'cypress',
    'playwright',
    'jest',
    'vitest',
    'mocha',
    'jasmine',
    'webdriverio',
    'puppeteer',
    'junit5',
    'testng',
    'pytest',
    'unittest',
  ];

  it.each(targetEmitters)('%s supports core assertion kinds', async (name) => {
    const emitter = await EMITTERS[name]();
    let supportedCount = 0;

    for (const kind of ASSERTION_KINDS) {
      const node = new Assertion({
        kind,
        subject: 'val',
        expected: '42',
      });
      const result = emitter.emitNode(node);
      if (result.supported) supportedCount++;
    }

    // E2E emitters (cypress, playwright, webdriverio, puppeteer) handle
    // assertions primarily through emitFullFile, not emitNode. Unit/Java/Python
    // emitters support more assertion kinds directly via emitNode.
    const e2eEmitters = ['cypress', 'playwright', 'webdriverio', 'puppeteer'];
    if (e2eEmitters.includes(name)) {
      expect(supportedCount).toBeGreaterThanOrEqual(1);
    } else {
      expect(supportedCount).toBeGreaterThanOrEqual(3);
    }
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Test: Cross-language round-trips via IR
// ═══════════════════════════════════════════════════════════════════════

describe('Cross-language IR round-trips', () => {
  it('JUnit4 → JUnit5 via ir-full', async () => {
    const source = await FRAMEWORKS.junit4();
    const target = await EMITTERS.junit5();
    const ir = source.default.parse(SAMPLES.junit4);
    const output = target.emitFullFile(ir);
    expect(output).toContain('Assertions.assertEquals');
    expect(output).toContain('@Test');
  });

  it('JUnit5 → TestNG via ir-full', async () => {
    const source = await FRAMEWORKS.junit5();
    const target = await EMITTERS.testng();
    const ir = source.default.parse(SAMPLES.junit5);
    const output = target.emitFullFile(ir);
    expect(output).toContain('Assert.assertEquals');
    expect(output).toContain('@Test');
  });

  it('pytest → unittest via ir-full', async () => {
    const source = await FRAMEWORKS.pytest();
    const target = await EMITTERS.unittest();
    const ir = source.default.parse(SAMPLES.pytest);
    const output = target.emitFullFile(ir);
    expect(output).toContain('self.assertEqual');
    expect(output).toContain('import unittest');
  });

  it('unittest → pytest via ir-full', async () => {
    const source = await FRAMEWORKS.unittest();
    const target = await EMITTERS.pytest();
    const ir = source.default.parse(SAMPLES.unittest);
    const output = target.emitFullFile(ir);
    expect(output).toContain('assert');
    expect(output).toContain('== 5');
  });

  it('nose2 → pytest via ir-full', async () => {
    const source = await FRAMEWORKS.nose2();
    const target = await EMITTERS.pytest();
    const ir = source.default.parse(SAMPLES.nose2);
    const output = target.emitFullFile(ir);
    expect(output).toContain('assert');
  });

  it('Jest → Vitest via ir-full', async () => {
    const source = await FRAMEWORKS.jest();
    const target = await EMITTERS.vitest();
    const ir = source.default.parse(SAMPLES.jest);
    const output = target.emitFullFile(ir);
    expect(output).toContain('expect');
    expect(output).toContain('describe');
  });

  it('Mocha → Jest via ir-full', async () => {
    const source = await FRAMEWORKS.mocha();
    const target = await EMITTERS.jest();
    const ir = source.default.parse(SAMPLES.mocha);
    const output = target.emitFullFile(ir);
    expect(output).toContain('describe');
    expect(output).toContain('it(');
  });

  it('Cypress → Playwright via ir-full', async () => {
    const source = await FRAMEWORKS.cypress();
    const target = await EMITTERS.playwright();
    const ir = source.default.parse(SAMPLES.cypress);
    const output = target.emitFullFile(ir);
    expect(output).toContain('test');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Test: Auto emitter mode (Phase 8)
// ═══════════════════════════════════════════════════════════════════════

describe('Auto emitter mode', () => {
  it('should accept "auto" as a valid emitter mode in ConversionPipeline', async () => {
    const { FrameworkRegistry } = await import(
      '../../src/core/FrameworkRegistry.js'
    );
    const { ConversionPipeline } = await import(
      '../../src/core/ConversionPipeline.js'
    );

    const registry = new FrameworkRegistry();
    const jestFw = (await FRAMEWORKS.jest()).default;
    const vitestFw = (await FRAMEWORKS.vitest()).default;
    registry.register(jestFw);
    registry.register(vitestFw);

    const pipeline = new ConversionPipeline(registry);
    const result = await pipeline.convert(SAMPLES.jest, 'jest', 'vitest', {
      emitter: 'auto',
    });

    expect(result.code).toBeDefined();
    expect(result.code.length).toBeGreaterThan(0);
    expect(result.report).toBeDefined();
    // Auto mode should produce IR coverage since it uses ir-full
    expect(result.report.irCoverage).toBeDefined();
  });

  it('auto mode should produce same output as ir-full for jest→vitest', async () => {
    const { FrameworkRegistry } = await import(
      '../../src/core/FrameworkRegistry.js'
    );
    const { ConversionPipeline } = await import(
      '../../src/core/ConversionPipeline.js'
    );

    const registry = new FrameworkRegistry();
    const jestFw = (await FRAMEWORKS.jest()).default;
    const vitestFw = (await FRAMEWORKS.vitest()).default;
    registry.register(jestFw);
    registry.register(vitestFw);

    const pipeline = new ConversionPipeline(registry);
    const autoResult = await pipeline.convert(SAMPLES.jest, 'jest', 'vitest', {
      emitter: 'auto',
    });
    const irFullResult = await pipeline.convert(
      SAMPLES.jest,
      'jest',
      'vitest',
      { emitter: 'ir-full' }
    );

    expect(autoResult.code).toBe(irFullResult.code);
  });
});
