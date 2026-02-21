/**
 * Scores conversion confidence for a converted IR tree.
 *
 * Walks the IR and counts converted vs unconvertible vs warning nodes,
 * applying weights by node type to produce a 0-100 confidence score.
 */

import { walkIR } from './ir.js';

const NODE_WEIGHTS = {
  TestSuite: 3,
  TestCase: 3,
  Hook: 3,
  Assertion: 2,
  Navigation: 2,
  MockCall: 2,
  ImportStatement: 1,
  RawCode: 1,
  Comment: 0,
  SharedVariable: 1,
  Modifier: 1,
  ParameterSet: 1,
  TestFile: 0,
};

export class ConfidenceScorer {
  /**
   * Score an IR tree and produce a confidence report.
   *
   * @param {import('./ir.js').TestFile} ir - Root IR node
   * @param {Object} [options]
   * @param {boolean} [options.experimentalIR] - Whether IR mode was active
   * @param {number} [options.totalSucceeded] - Number of IR nodes successfully emitted
   * @param {number} [options.totalAttempted] - Number of IR nodes attempted
   * @returns {Object} Report with confidence, counts, and details
   */
  score(ir, options = {}) {
    let totalWeight = 0;
    let convertedWeight = 0;
    let convertedCount = 0;
    let unconvertibleCount = 0;
    let warningCount = 0;
    const details = [];

    walkIR(ir, (node) => {
      const weight = NODE_WEIGHTS[node.type] ?? 0;
      if (weight === 0) return;

      totalWeight += weight;

      if (node.confidence === 'converted') {
        convertedWeight += weight;
        convertedCount++;
      } else if (node.confidence === 'unconvertible') {
        unconvertibleCount++;
        details.push({
          type: 'unconvertible',
          nodeType: node.type,
          line: node.sourceLocation ? node.sourceLocation.line : null,
          source: node.originalSource,
        });
      } else if (node.confidence === 'warning') {
        convertedWeight += weight;
        warningCount++;
        details.push({
          type: 'warning',
          nodeType: node.type,
          line: node.sourceLocation ? node.sourceLocation.line : null,
          source: node.originalSource,
        });
      }
    });

    let confidence =
      totalWeight > 0 ? Math.round((convertedWeight / totalWeight) * 100) : 100;

    // IR mode: ratio-based adjustment
    if (options.experimentalIR && options.totalAttempted > 0) {
      const irRatio = options.totalSucceeded / options.totalAttempted;
      confidence = Math.min(100, confidence + Math.round(irRatio * 5));
    }

    return {
      confidence,
      level: this.getLevel(confidence),
      converted: convertedCount,
      unconvertible: unconvertibleCount,
      warnings: warningCount,
      total: convertedCount + unconvertibleCount + warningCount,
      details,
    };
  }

  /**
   * Get the human-readable confidence level.
   * @param {number} confidence - 0-100
   * @returns {'high'|'medium'|'low'}
   */
  getLevel(confidence) {
    if (confidence >= 90) return 'high';
    if (confidence >= 70) return 'medium';
    return 'low';
  }
}
