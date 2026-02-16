/**
 * Shared helper for JUnit 4→JUnit 5 fixture tests.
 *
 * Reads input and expected files, runs conversion, asserts match.
 */
import fs from 'fs/promises';
import path from 'path';
import { ConversionPipeline } from '../../../src/core/ConversionPipeline.js';
import { FrameworkRegistry } from '../../../src/core/FrameworkRegistry.js';

let pipeline = null;

/**
 * Get or create the conversion pipeline.
 */
export async function getPipeline() {
  if (pipeline) return pipeline;

  const registry = new FrameworkRegistry();

  const { default: junit4Def } = await import('../../../src/languages/java/frameworks/junit4.js');
  const { default: junit5Def } = await import('../../../src/languages/java/frameworks/junit5.js');

  registry.register(junit4Def);
  registry.register(junit5Def);

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
 * @param {string} id - Test case ID (e.g., 'ANNOT-001')
 * @param {Object} [options]
 * @param {number} [options.minConfidence] - Minimum expected confidence (default: 90)
 * @param {string} [options.ext] - File extension (default: '.java')
 */
export async function runFixture(dir, id, options = {}) {
  const ext = options.ext || '.java';
  const minConfidence = options.minConfidence ?? 90;

  const inputPath = path.join(dir, `${id}.input${ext}`);
  const expectedPath = path.join(dir, `${id}.expected${ext}`);

  const input = await fs.readFile(inputPath, 'utf8');
  const expected = await fs.readFile(expectedPath, 'utf8');

  const pipe = await getPipeline();
  const result = await pipe.convert(input, 'junit4', 'junit5');

  // Normalize: trim trailing whitespace per line, normalize line endings
  const normalize = (s) =>
    s.split('\n').map(l => l.trimEnd()).join('\n').trim();

  expect(normalize(result.code)).toBe(normalize(expected));

  if (minConfidence > 0) {
    expect(result.report.confidence).toBeGreaterThanOrEqual(minConfidence);
  }

  return result;
}
