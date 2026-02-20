/**
 * The 5-stage conversion pipeline.
 *
 * Stages: Detect → Parse → Transform → Emit → Score
 *
 * For same-paradigm conversions (e.g., Jest→Vitest), the Transform
 * stage is a pass-through. For cross-paradigm conversions (e.g.,
 * pytest→unittest), the Transform stage applies structural changes.
 */

import { ConfidenceScorer } from './ConfidenceScorer.js';

export class ConversionPipeline {
  /**
   * @param {import('./FrameworkRegistry.js').FrameworkRegistry} registry
   */
  constructor(registry) {
    this.registry = registry;
    this.scorer = new ConfidenceScorer();
  }

  /**
   * Convert source code from one framework to another.
   *
   * @param {string} sourceCode - Source test file content
   * @param {string} sourceFrameworkName - Source framework name (e.g., 'jest')
   * @param {string} targetFrameworkName - Target framework name (e.g., 'vitest')
   * @param {Object} [options]
   * @param {string} [options.language] - Language hint for disambiguation
   * @returns {Promise<{code: string, report: Object}>}
   */
  async convert(
    sourceCode,
    sourceFrameworkName,
    targetFrameworkName,
    options = {}
  ) {
    const language = options.language || null;

    // 1. Detect — resolve framework definitions
    const sourceFw = this.registry.get(sourceFrameworkName, language);
    if (!sourceFw) {
      throw new Error(`Unknown source framework: '${sourceFrameworkName}'`);
    }

    const targetFw = this.registry.get(targetFrameworkName, language);
    if (!targetFw) {
      throw new Error(`Unknown target framework: '${targetFrameworkName}'`);
    }

    // Confirm source detection
    const detectionConfidence = sourceFw.detect(sourceCode);
    if (detectionConfidence === 0 && sourceCode.trim().length > 0) {
      throw new Error(
        `Source code does not appear to be ${sourceFrameworkName} (detection confidence: 0)`
      );
    }

    // 2. Parse — source framework parser produces IR
    const ir = sourceFw.parse(sourceCode);

    // 3. Transform — structural transforms for cross-paradigm
    const transformedIr = this.transform(ir, sourceFw, targetFw);

    // 4. Emit — target framework emitter produces code
    const code = targetFw.emit(transformedIr, sourceCode);

    // 5. Score — walk IR and compute confidence
    const report = this.scorer.score(transformedIr);

    return { code, report };
  }

  /**
   * Apply structural transforms when paradigms differ.
   * Currently a pass-through for same-paradigm conversions.
   *
   * @param {import('./ir.js').TestFile} ir
   * @param {Object} sourceFw - Source framework definition
   * @param {Object} targetFw - Target framework definition
   * @returns {import('./ir.js').TestFile}
   */
  transform(ir, sourceFw, targetFw) {
    if (sourceFw.paradigm === targetFw.paradigm) {
      return ir;
    }

    // Cross-paradigm structural transforms will be added here
    // when we implement pytest→unittest, RSpec→Minitest, etc.
    return ir;
  }
}
