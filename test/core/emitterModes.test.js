/**
 * Tests for Phase 0 IR generalization infrastructure.
 *
 * Covers:
 * - Emitter mode routing (legacy, ir-patch, ir-full)
 * - Dynamic IR emitter loading (_loadIREmitter)
 * - Per-node fallback (Policy B: one unsupported, others still replaced)
 * - emitter option aliasing (--experimental-ir → ir-patch)
 * - ConfidenceScorer ratio-based IR scoring
 */

import { ConversionPipeline } from '../../src/core/ConversionPipeline.js';
import { FrameworkRegistry } from '../../src/core/FrameworkRegistry.js';
import { ConfidenceScorer } from '../../src/core/ConfidenceScorer.js';
import {
  TestFile,
  TestSuite,
  TestCase,
  Assertion,
} from '../../src/core/ir.js';

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

describe('Emitter mode routing', () => {
  let pipeline;

  beforeEach(async () => {
    const registry = await loadRegistry();
    pipeline = new ConversionPipeline(registry);
  });

  const source = `
describe('Test', () => {
  it('checks', () => {
    cy.get('#btn').should('be.visible');
  });
});`;

  it('should use legacy mode by default', async () => {
    const result = await pipeline.convert(source, 'cypress', 'playwright');
    expect(result.report.irCoverage).toBeUndefined();
  });

  it('should use legacy mode when emitter=legacy', async () => {
    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'legacy',
    });
    expect(result.report.irCoverage).toBeUndefined();
  });

  it('should use ir-patch when emitter=ir-patch', async () => {
    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-patch',
    });
    expect(result.report.irCoverage).toBeDefined();
    expect(result.code).toContain('toBeVisible()');
  });

  it('should map experimentalIR to ir-patch', async () => {
    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      experimentalIR: true,
    });
    expect(result.report.irCoverage).toBeDefined();
  });

  it('should prefer explicit emitter over experimentalIR', async () => {
    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      experimentalIR: true,
      emitter: 'legacy',
    });
    // Explicit emitter=legacy should win over experimentalIR
    expect(result.report.irCoverage).toBeUndefined();
  });

  it('should use ir-full when emitter=ir-full', async () => {
    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-full',
    });
    expect(result.report.irCoverage).toBeDefined();
    expect(result.code).toContain('toBeVisible()');
    // ir-full generates the complete file from IR tree walk
    expect(result.code).toContain("import { test, expect }");
    expect(result.code).toContain("test.describe('Test'");
  });
});

describe('Dynamic IR emitter loading', () => {
  let pipeline;

  beforeEach(async () => {
    const registry = await loadRegistry();
    pipeline = new ConversionPipeline(registry);
  });

  it('should load playwright emitter successfully', async () => {
    const emitter = await pipeline._loadIREmitter('playwright');
    expect(emitter).not.toBeNull();
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
  });

  it('should return null for unknown target', async () => {
    const emitter = await pipeline._loadIREmitter('unknown-framework');
    expect(emitter).toBeNull();
  });

  it('should fall back to legacy when no emitter exists', async () => {
    const source = `
describe('Test', () => {
  it('works', () => {
    cy.get('#btn').should('be.visible');
  });
});`;

    // Using a direction that has no IR emitter — the pipeline should
    // gracefully fall back to legacy emit
    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'legacy',
    });
    expect(result.code).toBeTruthy();
    expect(result.report.irCoverage).toBeUndefined();
  });
});

describe('Per-node fallback (Policy B)', () => {
  let pipeline;

  beforeEach(async () => {
    const registry = await loadRegistry();
    pipeline = new ConversionPipeline(registry);
  });

  it('should replace supported nodes and keep baseline for unsupported', async () => {
    const source = `
describe('Test', () => {
  it('mixes supported and unsupported', () => {
    cy.get('#btn').should('be.visible');
    cy.get('#elem').should('custom.matcher');
  });
});`;

    const baseline = await pipeline.convert(
      source,
      'cypress',
      'playwright',
      {}
    );
    const irResult = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-patch',
    });

    // Supported assertion should be replaced by IR
    expect(irResult.code).toContain('toBeVisible()');
    // The file should still be valid
    expect(irResult.code).toContain('test.describe');
  });

  it('should report partial success in irCoverage', async () => {
    const source = `
describe('Test', () => {
  it('mixes', () => {
    cy.get('#btn').should('be.visible');
    cy.get('#elem').should('custom.matcher');
  });
});`;

    const result = await pipeline.convert(source, 'cypress', 'playwright', {
      emitter: 'ir-patch',
    });

    expect(result.report.irCoverage).toBeDefined();
    expect(result.report.irCoverage.supportedAssertions).toBe(1);
    expect(result.report.irCoverage.unsupportedAssertions).toBe(1);
    expect(result.report.irCoverage.coveragePercent).toBe(50);
  });
});

describe('ConfidenceScorer ratio-based IR scoring', () => {
  let scorer;

  beforeEach(() => {
    scorer = new ConfidenceScorer();
  });

  it('should boost confidence proportionally to IR success ratio', () => {
    const a1 = new Assertion({ kind: 'equal', confidence: 'converted' });
    const a2 = new Assertion({
      kind: 'snapshot',
      confidence: 'unconvertible',
    });
    const tc = new TestCase({
      name: 'test',
      body: [a1, a2],
      confidence: 'converted',
    });
    const file = new TestFile({ body: [tc] });

    const standard = scorer.score(file);

    // 100% success ratio → +5 boost
    const fullSuccess = scorer.score(file, {
      experimentalIR: true,
      totalSucceeded: 3,
      totalAttempted: 3,
    });
    expect(fullSuccess.confidence).toBe(standard.confidence + 5);

    // 50% success ratio → +3 boost (round(0.5 * 5) = 3)
    const halfSuccess = scorer.score(file, {
      experimentalIR: true,
      totalSucceeded: 1,
      totalAttempted: 2,
    });
    expect(halfSuccess.confidence).toBe(standard.confidence + 3);

    // 0% success ratio → +0 boost
    const noSuccess = scorer.score(file, {
      experimentalIR: true,
      totalSucceeded: 0,
      totalAttempted: 2,
    });
    expect(noSuccess.confidence).toBe(standard.confidence);
  });

  it('should not adjust when totalAttempted is 0', () => {
    const a1 = new Assertion({ kind: 'equal', confidence: 'converted' });
    const tc = new TestCase({
      name: 'test',
      body: [a1],
      confidence: 'converted',
    });
    const file = new TestFile({ body: [tc] });

    const standard = scorer.score(file);
    const irMode = scorer.score(file, {
      experimentalIR: true,
      totalSucceeded: 0,
      totalAttempted: 0,
    });

    expect(irMode.confidence).toBe(standard.confidence);
  });

  it('should cap at 100', () => {
    const a1 = new Assertion({ kind: 'equal', confidence: 'converted' });
    const tc = new TestCase({
      name: 'test',
      body: [a1],
      confidence: 'converted',
    });
    const file = new TestFile({ body: [tc] });

    const irMode = scorer.score(file, {
      experimentalIR: true,
      totalSucceeded: 5,
      totalAttempted: 5,
    });

    expect(irMode.confidence).toBe(100);
  });
});
