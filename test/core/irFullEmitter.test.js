/**
 * Tests for Phase 1-2: ir-full emitter and MockCall emission.
 *
 * Covers:
 * - Full file generation from IR tree walk
 * - test.describe blocks with names
 * - test() cases with async ({ page })
 * - Hook emission (beforeEach, afterEach, beforeAll, afterAll)
 * - Assertion and Navigation emission within test bodies
 * - RawCode and Comment pass-through
 * - Modifier support (.only, .skip)
 * - MockCall emission (networkIntercept → page.route)
 * - HAMLET-TODO for unsupported nodes
 * - Golden test: ir-full output matches expected fixture
 */

import { ConversionPipeline } from '../../src/core/ConversionPipeline.js';
import { FrameworkRegistry } from '../../src/core/FrameworkRegistry.js';
import {
  emitFullFile,
  emitAssertion,
  emitNavigation,
  emitMockCall,
  emitNode,
} from '../../src/core/emitters/playwright/irEmitter.js';
import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  Navigation,
  RawCode,
  Comment,
  MockCall,
  SharedVariable,
  Modifier,
} from '../../src/core/ir.js';
import fs from 'fs/promises';

async function loadRegistry() {
  const cypress = (
    await import('../../src/languages/javascript/frameworks/cypress.js')
  ).default;
  const playwright = (
    await import('../../src/languages/javascript/frameworks/playwright.js')
  ).default;
  const registry = new FrameworkRegistry();
  registry.register(cypress);
  registry.register(playwright);
  return registry;
}

describe('emitFullFile', () => {
  describe('basic structure', () => {
    it('should return null for non-TestFile input', () => {
      expect(emitFullFile({})).toBeNull();
      expect(emitFullFile(null)).toBeNull();
      expect(emitFullFile('string')).toBeNull();
    });

    it('should emit Playwright import for empty TestFile', () => {
      const ir = new TestFile({ language: 'javascript' });
      const result = emitFullFile(ir);
      expect(result).toContain("import { test, expect } from '@playwright/test'");
    });

    it('should emit test.describe with suite name', () => {
      const ir = new TestFile({
        body: [new TestSuite({ name: 'My Suite' })],
      });
      const result = emitFullFile(ir);
      expect(result).toContain("test.describe('My Suite', () => {");
    });

    it('should emit test() with case name and async page', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [new TestCase({ name: 'does something' })],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain("test('does something', async ({ page }) => {");
    });
  });

  describe('hooks', () => {
    it('should emit test.beforeEach with async page', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            hooks: [
              new Hook({
                hookType: 'beforeEach',
                body: [
                  new Navigation({ action: 'visit', url: '/app' }),
                ],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain('test.beforeEach(async ({ page }) => {');
      expect(result).toContain("await page.goto('/app');");
    });

    it('should emit all hook types', () => {
      const hooks = ['beforeAll', 'beforeEach', 'afterEach', 'afterAll'].map(
        (type) => new Hook({ hookType: type })
      );
      const ir = new TestFile({
        body: [new TestSuite({ name: 'Suite', hooks })],
      });
      const result = emitFullFile(ir);
      expect(result).toContain('test.beforeAll(async ({ page }) => {');
      expect(result).toContain('test.beforeEach(async ({ page }) => {');
      expect(result).toContain('test.afterEach(async ({ page }) => {');
      expect(result).toContain('test.afterAll(async ({ page }) => {');
    });

    it('should sort hooks in conventional order', () => {
      const hooks = [
        new Hook({ hookType: 'afterAll' }),
        new Hook({ hookType: 'beforeEach' }),
      ];
      const ir = new TestFile({
        body: [new TestSuite({ name: 'Suite', hooks })],
      });
      const result = emitFullFile(ir);
      const beforeIdx = result.indexOf('test.beforeEach');
      const afterIdx = result.indexOf('test.afterAll');
      expect(beforeIdx).toBeLessThan(afterIdx);
    });
  });

  describe('assertions and navigation in test bodies', () => {
    it('should emit assertions inside test cases', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'checks visibility',
                body: [
                  new Assertion({
                    kind: 'be.visible',
                    subject: '#btn',
                    isNegated: false,
                  }),
                ],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain(
        "await expect(page.locator('#btn')).toBeVisible();"
      );
    });

    it('should emit navigation inside test cases', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'navigates',
                body: [
                  new Navigation({ action: 'visit', url: '/home' }),
                  new Navigation({ action: 'goBack' }),
                  new Navigation({ action: 'reload' }),
                ],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain("await page.goto('/home');");
      expect(result).toContain('await page.goBack();');
      expect(result).toContain('await page.reload();');
    });
  });

  describe('pass-through nodes', () => {
    it('should pass through RawCode', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'test',
                body: [
                  new RawCode({ code: 'const data = getData();' }),
                ],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain('const data = getData();');
    });

    it('should pass through Comment', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'test',
                body: [new Comment({ text: '// important note' })],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain('// important note');
    });
  });

  describe('modifiers', () => {
    it('should emit test.describe.only for .only modifier', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Focused',
            modifiers: [new Modifier({ modifierType: 'only' })],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain("test.describe.only('Focused'");
    });

    it('should emit test.skip for .skip modifier on test', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'skipped',
                modifiers: [new Modifier({ modifierType: 'skip' })],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain("test.skip('skipped'");
    });
  });

  describe('HAMLET-TODO for unsupported nodes', () => {
    it('should emit TODO for MockCall', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'test',
                body: [
                  new MockCall({
                    kind: 'createMock',
                    originalSource: 'jest.fn()',
                  }),
                ],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('Unsupported mock kind');
    });

    it('should emit TODO for SharedVariable', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Suite',
            tests: [
              new TestCase({
                name: 'test',
                body: [new SharedVariable({ name: 'myVar' })],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain('HAMLET-TODO');
      expect(result).toContain('myVar');
    });
  });

  describe('nested suites', () => {
    it('should emit nested test.describe blocks', () => {
      const ir = new TestFile({
        body: [
          new TestSuite({
            name: 'Outer',
            tests: [
              new TestSuite({
                name: 'Inner',
                tests: [new TestCase({ name: 'nested test' })],
              }),
            ],
          }),
        ],
      });
      const result = emitFullFile(ir);
      expect(result).toContain("test.describe('Outer'");
      expect(result).toContain("test.describe('Inner'");
      expect(result).toContain("test('nested test'");
    });
  });
});

describe('emitMockCall', () => {
  it('should emit page.route with fulfill for networkIntercept with response', () => {
    const node = new MockCall({
      kind: 'networkIntercept',
      target: '/api/users',
      args: ['GET'],
      returnValue: '{ fixture: "users.json" }',
    });
    const result = emitMockCall(node);
    expect(result.supported).toBe(true);
    expect(result.code).toContain("page.route('/api/users'");
    expect(result.code).toContain('route.fulfill');
    expect(result.code).toContain('{ fixture: "users.json" }');
  });

  it('should emit page.route with continue for networkIntercept spy', () => {
    const node = new MockCall({
      kind: 'networkIntercept',
      target: '/api/data',
      args: ['GET'],
      returnValue: null,
    });
    const result = emitMockCall(node);
    expect(result.supported).toBe(true);
    expect(result.code).toContain("page.route('/api/data'");
    expect(result.code).toContain('route.continue()');
  });

  it('should emit TODO for createStub', () => {
    const node = new MockCall({
      kind: 'createStub',
      originalSource: 'cy.stub()',
    });
    const result = emitMockCall(node);
    expect(result.supported).toBe(false);
    expect(result.code).toContain('HAMLET-TODO');
    expect(result.code).toContain('cy.stub()');
  });

  it('should emit TODO for createSpy', () => {
    const node = new MockCall({
      kind: 'createSpy',
      originalSource: 'cy.spy()',
    });
    const result = emitMockCall(node);
    expect(result.supported).toBe(false);
    expect(result.code).toContain('HAMLET-TODO');
  });

  it('should emit TODO for fakeTimers', () => {
    const node = new MockCall({
      kind: 'fakeTimers',
      originalSource: 'cy.clock()',
    });
    const result = emitMockCall(node);
    expect(result.supported).toBe(false);
    expect(result.code).toContain('HAMLET-TODO');
    expect(result.code).toContain('page.clock');
  });

  it('should emit TODO for advanceTimers with ms', () => {
    const node = new MockCall({
      kind: 'advanceTimers',
      args: ['1000'],
      originalSource: 'cy.tick(1000)',
    });
    const result = emitMockCall(node);
    expect(result.supported).toBe(false);
    expect(result.code).toContain('HAMLET-TODO');
    expect(result.code).toContain('1000');
  });

  it('should return unsupported for non-MockCall input', () => {
    const result = emitMockCall({});
    expect(result.supported).toBe(false);
  });
});

describe('emitNode dispatches all node types', () => {
  it('should dispatch MockCall to emitMockCall', () => {
    const node = new MockCall({
      kind: 'networkIntercept',
      target: '/api/test',
      returnValue: '{ status: 200 }',
    });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toContain('page.route');
  });

  it('should dispatch Assertion to emitAssertion', () => {
    const node = new Assertion({
      kind: 'be.visible',
      subject: '#btn',
    });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toContain('toBeVisible');
  });

  it('should dispatch Navigation to emitNavigation', () => {
    const node = new Navigation({ action: 'visit', url: '/home' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toContain('page.goto');
  });

  it('should return unsupported for unknown node types', () => {
    const result = emitNode(new SharedVariable({ name: 'x' }));
    expect(result.supported).toBe(false);
  });
});

describe('emitFullFile with MockCall nodes', () => {
  it('should emit page.route for networkIntercept in test body', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'API Tests',
          tests: [
            new TestCase({
              name: 'intercepts API',
              body: [
                new MockCall({
                  kind: 'networkIntercept',
                  target: '/api/users',
                  args: ['GET'],
                  returnValue: '{ fixture: "users.json" }',
                }),
                new Navigation({ action: 'visit', url: '/users' }),
              ],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("page.route('/api/users'");
    expect(result).toContain('route.fulfill');
    expect(result).toContain("page.goto('/users')");
  });

  it('should emit TODO for unsupported mock kinds in test body', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [
            new TestCase({
              name: 'test',
              body: [
                new MockCall({
                  kind: 'createStub',
                  originalSource: 'cy.stub(win, "open")',
                }),
              ],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('HAMLET-TODO');
    expect(result).toContain('cy.stub()');
  });
});

describe('Golden tests: ir-full vs fixtures', () => {
  let pipeline;

  beforeEach(async () => {
    const registry = await loadRegistry();
    pipeline = new ConversionPipeline(registry);
  });

  it('should match assertions-basic ir-full expected output', async () => {
    const source = await fs.readFile(
      'test/fixtures/ir-pilot/assertions-basic.cy.js',
      'utf8'
    );
    const expected = await fs.readFile(
      'test/fixtures/ir-pilot/assertions-basic.ir.spec.js',
      'utf8'
    );

    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-full',
    });

    expect(result.code.trim()).toBe(expected.trim());
  });

  it('should produce valid output for all three emitter modes', async () => {
    const source = await fs.readFile(
      'test/fixtures/ir-pilot/assertions-basic.cy.js',
      'utf8'
    );

    const legacy = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'legacy',
    });
    const irPatch = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-patch',
    });
    const irFull = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-full',
    });

    // All three should produce valid, non-empty output
    expect(legacy.code.length).toBeGreaterThan(0);
    expect(irPatch.code.length).toBeGreaterThan(0);
    expect(irFull.code.length).toBeGreaterThan(0);

    // All should contain Playwright import
    expect(legacy.code).toContain("from '@playwright/test'");
    expect(irPatch.code).toContain("from '@playwright/test'");
    expect(irFull.code).toContain("from '@playwright/test'");

    // All should have test structure
    expect(legacy.code).toContain('test.describe');
    expect(irPatch.code).toContain('test.describe');
    expect(irFull.code).toContain('test.describe');

    // ir-full should have irCoverage in report
    expect(legacy.report.irCoverage).toBeUndefined();
    expect(irPatch.report.irCoverage).toBeDefined();
    expect(irFull.report.irCoverage).toBeDefined();
  });
});
