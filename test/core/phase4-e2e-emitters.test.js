/**
 * Phase 4 tests: E2E target emitters (Cypress, WebdriverIO, Puppeteer)
 * and Playwright parser enhancement.
 *
 * Covers:
 * - Playwright parser: nested IR with assertion/navigation/mock extraction
 * - Cypress ir-full emitter: cy.get().should(), cy.visit(), cy.intercept()
 * - WebdriverIO ir-full emitter: $(), browser.url(), await expect()
 * - Puppeteer ir-full emitter: page.$(), page.goto(), browser boilerplate
 * - Dynamic emitter loading for all four E2E targets
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  Navigation,
  MockCall,
  RawCode,
  Comment,
  ImportStatement,
  SharedVariable,
  Modifier,
  walkIR,
} from '../../src/core/ir.js';

// ═══════════════════════════════════════════════════════════════════════
// Playwright parser: nested IR
// ═══════════════════════════════════════════════════════════════════════

describe('Playwright parser — nested IR', () => {
  let parse;

  beforeEach(async () => {
    const pw = (
      await import('../../src/languages/javascript/frameworks/playwright.js')
    ).default;
    parse = pw.parse;
  });

  it('should produce nested TestSuite with TestCase children', () => {
    const ir = parse(`
import { test, expect } from '@playwright/test';

test.describe('Login', () => {
  test('should login', async ({ page }) => {
    await page.goto('/login');
  });
});
    `);
    expect(ir).toBeInstanceOf(TestFile);
    expect(ir.imports.length).toBe(1);

    const suite = ir.body[0];
    expect(suite).toBeInstanceOf(TestSuite);
    expect(suite.name).toBe('Login');
    expect(suite.tests.length).toBe(1);
    expect(suite.tests[0]).toBeInstanceOf(TestCase);
    expect(suite.tests[0].name).toBe('should login');
  });

  it('should nest hooks inside TestSuite.hooks[]', () => {
    const ir = parse(`
test.describe('Setup', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });
  test('works', async ({ page }) => {});
});
    `);
    const suite = ir.body[0];
    expect(suite.hooks.length).toBe(1);
    expect(suite.hooks[0]).toBeInstanceOf(Hook);
    expect(suite.hooks[0].hookType).toBe('beforeEach');
  });

  it('should extract navigation from body', () => {
    const ir = parse(`
test.describe('Nav', () => {
  test('navigates', async ({ page }) => {
    await page.goto('/home');
    await page.goBack();
    await page.reload();
  });
});
    `);
    const body = ir.body[0].tests[0].body;
    expect(body.length).toBe(3);
    expect(body[0]).toBeInstanceOf(Navigation);
    expect(body[0].action).toBe('visit');
    expect(body[0].url).toBe('/home');
    expect(body[1]).toBeInstanceOf(Navigation);
    expect(body[1].action).toBe('goBack');
    expect(body[2]).toBeInstanceOf(Navigation);
    expect(body[2].action).toBe('reload');
  });

  it('should extract assertions with subject', () => {
    const ir = parse(`
test.describe('Assertions', () => {
  test('checks visibility', async ({ page }) => {
    await expect(page.locator('#btn')).toBeVisible();
  });
});
    `);
    const assertion = ir.body[0].tests[0].body[0];
    expect(assertion).toBeInstanceOf(Assertion);
    expect(assertion.kind).toBe('be.visible');
    expect(assertion.subject).toBe('#btn');
  });

  it('should extract page-level assertions', () => {
    const ir = parse(`
test.describe('Page', () => {
  test('checks URL', async ({ page }) => {
    await expect(page).toHaveURL('https://example.com');
  });
});
    `);
    const assertion = ir.body[0].tests[0].body[0];
    expect(assertion).toBeInstanceOf(Assertion);
    expect(assertion.kind).toBe('url.equal');
  });

  it('should extract MockCall for page.route', () => {
    const ir = parse(`
test.describe('Mocks', () => {
  test('intercepts', async ({ page }) => {
    await page.route('/api/users', route => route.fulfill({ body: '[]' }));
  });
});
    `);
    const mock = ir.body[0].tests[0].body[0];
    expect(mock).toBeInstanceOf(MockCall);
    expect(mock.kind).toBe('networkIntercept');
    expect(mock.target).toBe('/api/users');
  });

  it('should handle modifiers (.only, .skip)', () => {
    const ir = parse(`
test.describe.only('focused', () => {
  test.skip('skipped', async ({ page }) => {});
});
    `);
    const suite = ir.body[0];
    expect(suite.modifiers[0].modifierType).toBe('only');
    const tc = suite.tests[0];
    expect(tc.modifiers[0].modifierType).toBe('skip');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Cypress ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('Cypress ir-full emitter', () => {
  let emitFullFile, emitAssertion, emitNavigation, emitMockCall;

  beforeEach(async () => {
    const mod = await import('../../src/core/emitters/cypress/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitAssertion = mod.emitAssertion;
    emitNavigation = mod.emitNavigation;
    emitMockCall = mod.emitMockCall;
  });

  it('should emit describe/it without framework imports', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Login',
          tests: [new TestCase({ name: 'works' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("describe('Login', () => {");
    expect(result).toContain("it('works', () => {");
    expect(result).not.toContain('import');
    expect(result).not.toContain('async');
  });

  it('should emit cy.get().should() for locator assertions', () => {
    const result = emitAssertion(
      new Assertion({ kind: 'be.visible', subject: '#btn' })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe("cy.get('#btn').should('be.visible')");
  });

  it('should emit cy.visit() for navigation', () => {
    const result = emitNavigation(
      new Navigation({ action: 'visit', url: '/home' })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe("cy.visit('/home')");
  });

  it('should emit cy.go("back") for goBack', () => {
    const result = emitNavigation(new Navigation({ action: 'goBack' }));
    expect(result.supported).toBe(true);
    expect(result.code).toBe("cy.go('back')");
  });

  it('should emit cy.intercept() for networkIntercept', () => {
    const result = emitMockCall(
      new MockCall({
        kind: 'networkIntercept',
        target: '/api/users',
        args: ['GET'],
        returnValue: '{ fixture: "users.json" }',
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toContain("cy.intercept('GET', '/api/users'");
  });

  it('should use before/after for hooks', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          hooks: [
            new Hook({ hookType: 'beforeAll' }),
            new Hook({ hookType: 'afterAll' }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('before(() => {');
    expect(result).toContain('after(() => {');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// WebdriverIO ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('WebdriverIO ir-full emitter', () => {
  let emitFullFile, emitAssertion, emitNavigation;

  beforeEach(async () => {
    const mod =
      await import('../../src/core/emitters/webdriverio/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitAssertion = mod.emitAssertion;
    emitNavigation = mod.emitNavigation;
  });

  it('should emit describe/it with async callbacks', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Login',
          tests: [new TestCase({ name: 'works' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("describe('Login', () => {");
    expect(result).toContain("it('works', async () => {");
  });

  it('should emit WDIO locator assertions with $() syntax', () => {
    const result = emitAssertion(
      new Assertion({ kind: 'be.visible', subject: '#btn' })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe("await expect($('#btn')).toBeDisplayed()");
  });

  it('should emit browser.url() for navigation', () => {
    const result = emitNavigation(
      new Navigation({ action: 'visit', url: '/home' })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe("await browser.url('/home')");
  });

  it('should emit browser-level assertions for URL', () => {
    const result = emitAssertion(
      new Assertion({
        kind: 'url.include',
        subject: 'page',
        expected: "'/login'",
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toContain('expect(browser)');
    expect(result.code).toContain("toHaveUrlContaining('/login')");
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Puppeteer ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('Puppeteer ir-full emitter', () => {
  let emitFullFile, emitAssertion, emitNavigation;

  beforeEach(async () => {
    const mod = await import('../../src/core/emitters/puppeteer/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitAssertion = mod.emitAssertion;
    emitNavigation = mod.emitNavigation;
  });

  it('should emit puppeteer require and browser boilerplate', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Login',
          tests: [new TestCase({ name: 'works' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("require('puppeteer')");
    expect(result).toContain('let browser, page;');
    expect(result).toContain('puppeteer.launch()');
    expect(result).toContain('browser.newPage()');
    expect(result).toContain('browser.close()');
  });

  it('should emit page.goto() for navigation', () => {
    const result = emitNavigation(
      new Navigation({ action: 'visit', url: '/home' })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe("await page.goto('/home')");
  });

  it('should emit page.$eval assertions for text', () => {
    const result = emitAssertion(
      new Assertion({
        kind: 'have.text',
        subject: '#title',
        expected: "'Hello'",
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toContain("page.$eval('#title'");
    expect(result.code).toContain("toBe('Hello')");
  });

  it('should emit page.url() for URL assertions', () => {
    const result = emitAssertion(
      new Assertion({
        kind: 'url.equal',
        subject: 'page',
        expected: "'https://example.com'",
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toContain('page.url()');
    expect(result.code).toContain("toBe('https://example.com')");
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Dynamic emitter loading for E2E targets
// ═══════════════════════════════════════════════════════════════════════

describe('Dynamic emitter loading — E2E targets', () => {
  it('should load cypress emitter', async () => {
    const emitter =
      await import('../../src/core/emitters/cypress/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should load webdriverio emitter', async () => {
    const emitter =
      await import('../../src/core/emitters/webdriverio/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should load puppeteer emitter', async () => {
    const emitter =
      await import('../../src/core/emitters/puppeteer/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });
});
