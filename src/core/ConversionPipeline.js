/**
 * The 5-stage conversion pipeline.
 *
 * Stages: Detect → Parse → Transform → Emit → Score
 *
 * For same-paradigm conversions (e.g., Jest→Vitest), the Transform
 * stage is a pass-through. For cross-paradigm conversions (e.g.,
 * pytest→unittest), the Transform stage applies structural changes.
 *
 * Emitter modes (--emitter=<mode>):
 *   legacy   — regex-based emit (default)
 *   ir-patch — IR emits per-node, patches into baseline (Policy B)
 *   ir-full  — IR generates complete file from tree walk
 *   auto     — tries ir-full, falls back to ir-patch, then legacy
 */

import { ConfidenceScorer } from './ConfidenceScorer.js';
import { Assertion, Navigation, MockCall, walkIR } from './ir.js';
import { computeIRCoverage } from './ir/coverage.js';

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
   * @param {boolean} [options.experimentalIR] - Alias for emitter='ir-patch'
   * @param {'legacy'|'ir-patch'|'ir-full'|'auto'} [options.emitter] - Emitter mode
   * @returns {Promise<{code: string, report: Object}>}
   */
  async convert(
    sourceCode,
    sourceFrameworkName,
    targetFrameworkName,
    options = {}
  ) {
    const language = options.language || null;

    // Resolve emitter mode: explicit > experimentalIR flag > legacy
    const emitterMode =
      options.emitter || (options.experimentalIR ? 'ir-patch' : 'legacy');

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

    // 4. Emit — route to appropriate emitter
    let code;
    let irEmissionMeta = null;

    if (
      emitterMode === 'ir-full' ||
      emitterMode === 'ir-patch' ||
      emitterMode === 'auto'
    ) {
      const irEmitter = await this._loadIREmitter(targetFrameworkName);

      if ((emitterMode === 'ir-full' || emitterMode === 'auto') && irEmitter) {
        const fullResult = this.emitWithIRFull(transformedIr, irEmitter);
        if (fullResult !== null) {
          code = fullResult;
          const nodes = [];
          walkIR(transformedIr, (node) => {
            if (
              node instanceof Assertion ||
              node instanceof Navigation ||
              node instanceof MockCall
            ) {
              nodes.push(node);
            }
          });
          irEmissionMeta = {
            succeeded: true,
            assertionCount: nodes.length,
            totalSucceeded: nodes.length,
            totalAttempted: nodes.length,
          };
        }
      }

      // Fallback: ir-full returned null or mode is ir-patch
      if (!code) {
        if (irEmitter) {
          const irResult = await this.emitWithIRPatch(
            transformedIr,
            sourceCode,
            targetFw,
            irEmitter
          );
          code = irResult.code;
          irEmissionMeta = {
            succeeded: irResult.succeeded,
            assertionCount: irResult.assertionCount,
            totalSucceeded: irResult.totalSucceeded,
            totalAttempted: irResult.totalAttempted,
          };
        } else {
          // No IR emitter available — fall back to legacy
          code = targetFw.emit(transformedIr, sourceCode);
        }
      }
    } else {
      // Legacy mode
      code = targetFw.emit(transformedIr, sourceCode);
    }

    // 5. Score — walk IR and compute confidence
    const isIRActive = emitterMode !== 'legacy' && irEmissionMeta !== null;
    const report = this.scorer.score(transformedIr, {
      experimentalIR: isIRActive,
      totalSucceeded: irEmissionMeta?.totalSucceeded ?? 0,
      totalAttempted: irEmissionMeta?.totalAttempted ?? 0,
    });

    // 6. IR coverage (when any IR mode is active)
    if (isIRActive) {
      report.irCoverage = computeIRCoverage(transformedIr);
    }

    return { code, report };
  }

  /**
   * Load the IR emitter module for a target framework.
   * Returns null if no emitter exists for the target.
   *
   * @param {string} targetName - Target framework name
   * @returns {Promise<Object|null>}
   */
  async _loadIREmitter(targetName) {
    try {
      return await import(`./emitters/${targetName}/irEmitter.js`);
    } catch {
      return null;
    }
  }

  /**
   * Emit using IR per-node, patching into regex baseline.
   * Policy B: per-node fallback — unsupported nodes keep baseline,
   * supported nodes get IR-emitted replacement.
   *
   * @param {import('./ir.js').TestFile} ir
   * @param {string} sourceCode
   * @param {Object} targetFw - Target framework definition
   * @param {Object} irEmitter - IR emitter module with emitNode/matchesBaseline
   * @returns {Promise<{code: string, succeeded: boolean, assertionCount: number, totalSucceeded: number, totalAttempted: number}>}
   */
  async emitWithIRPatch(ir, sourceCode, targetFw, irEmitter) {
    const baselineCode = targetFw.emit(ir, sourceCode);

    // Collect all IR-emittable nodes
    const irNodes = [];
    walkIR(ir, (node) => {
      if (
        node instanceof Assertion ||
        node instanceof Navigation ||
        node instanceof MockCall
      ) {
        irNodes.push(node);
      }
    });

    if (irNodes.length === 0) {
      return {
        code: baselineCode,
        succeeded: true,
        assertionCount: 0,
        totalSucceeded: 0,
        totalAttempted: 0,
      };
    }

    // Attempt IR emission for each node
    const results = irNodes.map((node) => irEmitter.emitNode(node));

    // Policy B: per-node fallback — replace only supported nodes
    let totalSucceeded = 0;
    const totalAttempted = irNodes.length;

    let irCode = baselineCode;
    for (let i = 0; i < irNodes.length; i++) {
      const node = irNodes[i];
      const emitted = results[i];

      if (!emitted.supported) continue;
      if (!node.originalSource) continue;

      totalSucceeded++;
      const lines = irCode.split('\n');
      let replaced = false;
      for (let j = 0; j < lines.length; j++) {
        const trimmedLine = lines[j].trim();
        if (
          !replaced &&
          trimmedLine.length > 0 &&
          irEmitter.matchesBaseline(trimmedLine, node)
        ) {
          const indent = lines[j].match(/^(\s*)/)[1];
          const semi = lines[j].trimEnd().endsWith(';') ? ';' : '';
          lines[j] = indent + emitted.code + semi;
          replaced = true;
        }
      }
      irCode = lines.join('\n');
    }

    return {
      code: irCode,
      succeeded: totalSucceeded === totalAttempted,
      assertionCount: totalAttempted,
      totalSucceeded,
      totalAttempted,
    };
  }

  /**
   * Generate complete output file from IR tree walk.
   * Returns null when not yet implemented — pipeline falls back to ir-patch.
   *
   * @param {import('./ir.js').TestFile} ir
   * @param {Object} irEmitter - IR emitter module with emitFullFile
   * @returns {string|null}
   */
  emitWithIRFull(ir, irEmitter) {
    if (typeof irEmitter.emitFullFile !== 'function') return null;
    return irEmitter.emitFullFile(ir);
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
