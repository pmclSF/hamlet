/**
 * Shared helper for Jest→Jasmine fixture tests.
 *
 * Reads input and expected files, runs conversion, asserts match.
 * This file is a test utility, not a test itself (no .test.js suffix).
 */
import fs from 'fs/promises';
import path from 'path';
import { ConversionPipeline } from '../../../src/core/ConversionPipeline.js';
import { FrameworkRegistry } from '../../../src/core/FrameworkRegistry.js';

let pipeline = null;

/**
 * Get or create the conversion pipeline.
 * Lazily initialized so framework definitions only need to exist when tests run.
 */
export async function getPipeline() {
  if (pipeline) return pipeline;

  const registry = new FrameworkRegistry();

  const { default: jestDef } = await import('../../../src/languages/javascript/frameworks/jest.js');
  const { default: jasmineDef } = await import('../../../src/languages/javascript/frameworks/jasmine.js');

  registry.register(jestDef);
  registry.register(jasmineDef);

  pipeline = new ConversionPipeline(registry);
  return pipeline;
}

/**
 * Reset the pipeline (for test isolation if needed).
 */
export function resetPipeline() {
  pipeline = null;
}

/**
 * Run a fixture test: read input, convert, compare to expected.
 *
 * @param {string} dir - Directory containing the fixture files
 * @param {string} id - Test case ID (e.g., 'STRUCTURE-001')
 * @param {Object} [options]
 * @param {number} [options.minConfidence] - Minimum expected confidence (default: 90)
 * @param {string} [options.ext] - File extension (default: '.js')
 */
export async function runFixture(dir, id, options = {}) {
  const ext = options.ext || '.js';
  const minConfidence = options.minConfidence ?? 90;

  const inputPath = path.join(dir, `${id}.input${ext}`);
  const expectedPath = path.join(dir, `${id}.expected${ext}`);

  const input = await fs.readFile(inputPath, 'utf8');
  const expected = await fs.readFile(expectedPath, 'utf8');

  const pipe = await getPipeline();
  const result = await pipe.convert(input, 'jest', 'jasmine');

  // Normalize: trim trailing whitespace per line, normalize line endings
  const normalize = (s) =>
    s.split('\n').map(l => l.trimEnd()).join('\n').trim();

  expect(normalize(result.code)).toBe(normalize(expected));

  if (minConfidence > 0) {
    expect(result.report.confidence).toBeGreaterThanOrEqual(minConfidence);
  }

  return result;
}
