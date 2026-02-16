/**
 * Estimates migration complexity without performing actual conversion.
 *
 * Runs scanner + classifier + dependency graph.
 * Returns structured data (CLI formats it).
 * Read-only: does NOT modify files, does NOT create .hamlet/.
 */

import fs from "fs/promises";
import path from "path";
import { Scanner } from "./Scanner.js";
import { FileClassifier } from "./FileClassifier.js";
import { DependencyGraphBuilder } from "./DependencyGraphBuilder.js";

const COMPLEXITY_PATTERNS = {
  jest: {
    high: [/jest\.mock\(/, /jest\.spyOn\(/, /jest\.requireActual\(/],
    medium: [/jest\.fn\(/, /jest\.useFakeTimers\(/, /jest\.setTimeout\(/],
    low: [/describe\(/, /it\(/, /test\(/, /expect\(/],
  },
  cypress: {
    high: [/cy\.intercept\(/, /cy\.stub\(/, /Cypress\.Commands\.add\(/],
    medium: [/cy\.fixture\(/, /cy\.wrap\(/, /cy\.task\(/],
    low: [/cy\.visit\(/, /cy\.get\(/, /cy\.contains\(/],
  },
};

export class MigrationEstimator {
  constructor() {
    this.scanner = new Scanner();
    this.classifier = new FileClassifier();
    this.graphBuilder = new DependencyGraphBuilder();
  }

  /**
   * Estimate migration complexity for a project.
   *
   * @param {string} rootDir - Project root directory
   * @param {Object} options
   * @param {string} options.from - Source framework
   * @param {string} options.to - Target framework
   * @returns {Promise<Object>} Structured estimation data
   */
  async estimate(rootDir, options) {
    const { from, to } = options;
    const resolvedRoot = path.resolve(rootDir);

    // Scan
    const scanned = await this.scanner.scan(resolvedRoot, {
      include: ["*.js", "*.ts", "*.jsx", "*.tsx", "*.mjs"],
    });

    // Read and classify
    const files = [];
    for (const entry of scanned) {
      let content;
      try {
        content = await fs.readFile(entry.path, "utf8");
      } catch {
        continue;
      }

      const classification = this.classifier.classify(
        entry.relativePath,
        content,
      );
      const complexity = this._estimateFileComplexity(content, from);

      files.push({
        ...entry,
        content,
        classification,
        complexity,
      });
    }

    // Build dependency graph
    const graph = this.graphBuilder.build(files);

    // Aggregate results
    const fileEstimates = files.map((f) => ({
      path: f.relativePath,
      type: f.classification.type,
      framework: f.classification.framework,
      predictedConfidence: this._predictConfidence(f.complexity),
      complexity: f.complexity,
    }));

    const high = fileEstimates.filter(
      (f) => f.predictedConfidence >= 90,
    ).length;
    const medium = fileEstimates.filter(
      (f) => f.predictedConfidence >= 70 && f.predictedConfidence < 90,
    ).length;
    const low = fileEstimates.filter(
      (f) => f.predictedConfidence > 0 && f.predictedConfidence < 70,
    ).length;

    // Identify top blockers
    const blockers = this._identifyBlockers(files, from);

    return {
      summary: {
        totalFiles: files.length,
        testFiles: fileEstimates.filter((f) => f.type === "test").length,
        helperFiles: fileEstimates.filter((f) => f.type === "helper").length,
        configFiles: fileEstimates.filter((f) => f.type === "config").length,
        otherFiles: fileEstimates.filter(
          (f) => !["test", "helper", "config"].includes(f.type),
        ).length,
        predictedHigh: high,
        predictedMedium: medium,
        predictedLow: low,
        circularDependencies: graph.cycles.length,
      },
      files: fileEstimates,
      blockers,
      estimatedEffort: this._estimateEffort(fileEstimates),
      from,
      to,
    };
  }

  /**
   * Estimate complexity for a single file.
   *
   * @param {string} content
   * @param {string} framework
   * @returns {{highPatterns: number, mediumPatterns: number, lowPatterns: number}}
   */
  _estimateFileComplexity(content, framework) {
    const patterns = COMPLEXITY_PATTERNS[framework] || {};
    let highPatterns = 0;
    let mediumPatterns = 0;
    let lowPatterns = 0;

    for (const pattern of patterns.high || []) {
      const matches = content.match(new RegExp(pattern.source, "g"));
      if (matches) highPatterns += matches.length;
    }

    for (const pattern of patterns.medium || []) {
      const matches = content.match(new RegExp(pattern.source, "g"));
      if (matches) mediumPatterns += matches.length;
    }

    for (const pattern of patterns.low || []) {
      const matches = content.match(new RegExp(pattern.source, "g"));
      if (matches) lowPatterns += matches.length;
    }

    return { highPatterns, mediumPatterns, lowPatterns };
  }

  /**
   * Predict confidence based on complexity.
   *
   * @param {{highPatterns: number, mediumPatterns: number, lowPatterns: number}} complexity
   * @returns {number} Predicted confidence 0-100
   */
  _predictConfidence(complexity) {
    if (complexity.highPatterns > 3) return 50;
    if (complexity.highPatterns > 0) return 70;
    if (complexity.mediumPatterns > 3) return 80;
    if (complexity.mediumPatterns > 0) return 85;
    return 95;
  }

  /**
   * Identify the most common patterns that would lower confidence.
   *
   * @param {Array} files
   * @param {string} framework
   * @returns {Array<{pattern: string, count: number, impact: string}>}
   */
  _identifyBlockers(files, framework) {
    const patterns = COMPLEXITY_PATTERNS[framework] || {};
    const counts = new Map();

    for (const file of files) {
      for (const pattern of patterns.high || []) {
        const matches = file.content.match(new RegExp(pattern.source, "g"));
        if (matches) {
          const key = pattern.source;
          counts.set(key, (counts.get(key) || 0) + matches.length);
        }
      }
    }

    return Array.from(counts.entries())
      .map(([pattern, count]) => ({
        pattern: pattern.replace(/\\\(/g, "(").replace(/\\\./g, "."),
        count,
        impact: "high",
      }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5);
  }

  /**
   * Estimate manual effort based on predicted confidence.
   *
   * @param {Array} fileEstimates
   * @returns {{lowConfidenceFiles: number, estimatedManualMinutes: number, description: string}}
   */
  _estimateEffort(fileEstimates) {
    const lowFiles = fileEstimates.filter((f) => f.predictedConfidence < 70);
    const mediumFiles = fileEstimates.filter(
      (f) => f.predictedConfidence >= 70 && f.predictedConfidence < 90,
    );

    // Rough estimate: 15 min per low-confidence file, 5 min per medium
    const minutes = lowFiles.length * 15 + mediumFiles.length * 5;

    let description;
    if (minutes === 0) {
      description = "Fully automated — no manual intervention expected";
    } else if (minutes < 30) {
      description = "Minimal manual effort expected";
    } else if (minutes < 120) {
      description = "Moderate manual effort expected";
    } else {
      description = "Significant manual effort expected";
    }

    return {
      lowConfidenceFiles: lowFiles.length,
      mediumConfidenceFiles: mediumFiles.length,
      estimatedManualMinutes: minutes,
      description,
    };
  }
}
