/**
 * Analyzes a project directory to detect testing frameworks and classify files.
 *
 * Composes Scanner + FileClassifier + FrameworkDetector to produce a
 * structured analysis report (JSON contract: docs/schema/analysis.schema.json).
 * Read-only: does NOT modify files or create output.
 */

import fs from 'fs/promises';
import path from 'path';
import { createRequire } from 'module';
import { Scanner } from './Scanner.js';
import { FileClassifier } from './FileClassifier.js';
import { FrameworkDetector } from './FrameworkDetector.js';
import { ConverterFactory } from './ConverterFactory.js';

const __require = createRequire(import.meta.url);
const hamletVersion = __require('../../package.json').version;

export class ProjectAnalyzer {
  constructor() {
    this.scanner = new Scanner();
    this.classifier = new FileClassifier();
  }

  /**
   * Analyze a project directory.
   *
   * @param {string} rootDir - Directory to scan
   * @param {Object} [options]
   * @param {number} [options.maxFiles=5000] - Maximum files to process
   * @param {string[]} [options.include] - Include glob patterns
   * @param {string[]} [options.exclude] - Exclude glob patterns
   * @returns {Promise<Object>} Analysis report matching analysis.schema.json
   */
  async analyze(rootDir, options = {}) {
    const resolvedRoot = path.resolve(rootDir);
    const maxFiles = options.maxFiles || 5000;

    // Scan
    const scanned = await this.scanner.scan(resolvedRoot, {
      include: options.include || [],
      exclude: options.exclude || [],
    });

    // Cap to maxFiles
    const capped = scanned.slice(0, maxFiles);

    // Classify each file
    const files = [];
    const frameworkSet = new Set();

    for (const entry of capped) {
      let content;
      try {
        content = await fs.readFile(entry.path, 'utf8');
      } catch {
        continue;
      }

      const classification = this.classifier.classify(
        entry.relativePath,
        content
      );

      // Get candidates from FrameworkDetector
      let candidates = [];
      try {
        const detection = FrameworkDetector.detectFromContent(content);
        candidates = Object.entries(detection.scores)
          .filter(([, score]) => score > 0)
          .map(([framework, score]) => ({ framework, score }))
          .sort((a, b) => b.score - a.score);
      } catch {
        // Detection may fail on non-code files
      }

      if (classification.framework) {
        frameworkSet.add(classification.framework);
      }

      files.push({
        path: entry.relativePath,
        type: classification.type,
        framework: classification.framework,
        candidates,
        confidence: classification.confidence,
        warnings: [],
      });
    }

    // Deterministic sort by path using < > (not localeCompare)
    files.sort((a, b) => (a.path < b.path ? -1 : a.path > b.path ? 1 : 0));

    // Build directions from detected frameworks
    const detectedFrameworks = [...frameworkSet].sort();
    const allDirections = ConverterFactory.getSupportedConversions();
    const directionsSupported = [];

    for (const dir of allDirections) {
      const [from, to] = dir.split('-');
      if (frameworkSet.has(from)) {
        directionsSupported.push({
          from,
          to,
          pipelineBacked: ConverterFactory.isPipelineBacked(from, to),
        });
      }
    }

    // Compute average confidence (only from non-zero values)
    const nonZero = files.filter((f) => f.confidence > 0);
    const confidenceAvg =
      nonZero.length > 0
        ? Math.round(
            (nonZero.reduce((sum, f) => sum + f.confidence, 0) /
              nonZero.length) *
              100
          ) / 100
        : 0;

    const testFileCount = files.filter((f) => f.type === 'test').length;

    return {
      schemaVersion: '1.0.0',
      meta: {
        hamletVersion,
        nodeVersion: process.version,
        generatedAt: new Date().toISOString(),
        root: resolvedRoot,
      },
      summary: {
        fileCount: files.length,
        testFileCount,
        frameworksDetected: detectedFrameworks,
        directionsSupported,
        confidenceAvg,
      },
      files,
    };
  }
}
