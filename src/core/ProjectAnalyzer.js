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

      // Compute meaningful confidence from detection data + content complexity
      const confidence = this._computeConfidence(
        classification,
        candidates,
        content
      );

      files.push({
        path: entry.relativePath,
        type: classification.type,
        framework: classification.framework,
        candidates,
        confidence,
        warnings: [],
      });
    }

    // Drop files with no relevance to test conversion
    const nonCodeExts = new Set([
      '.json',
      '.xml',
      '.csv',
      '.sql',
      '.txt',
      '.html',
      '.htm',
      '.css',
      '.scss',
      '.less',
      '.sass',
      '.svg',
      '.png',
      '.jpg',
      '.jpeg',
      '.gif',
      '.ico',
      '.webp',
      '.woff',
      '.woff2',
      '.ttf',
      '.eot',
      '.mp3',
      '.mp4',
      '.webm',
      '.pdf',
      '.zip',
      '.tar',
      '.gz',
      '.lock',
      '.map',
      '.snap',
      '.md',
      '.mdx',
      '.rst',
      '.yml',
      '.yaml',
      '.toml',
      '.ini',
      '.cfg',
      '.env',
      '.gitignore',
      '.dockerignore',
      '.editorconfig',
      '.prettierrc',
      '.eslintignore',
      '.npmignore',
      '.npmrc',
      '.nvmrc',
      '.sh',
      '.bat',
      '.cmd',
      '.ps1',
    ]);
    const relevant = files.filter((f) => {
      if (f.type === 'unknown' && f.confidence === 0) return false;
      // Keep config files only if they belong to a test framework
      if (f.type === 'config' && !f.framework) return false;
      // Drop non-code files — never need conversion
      const ext = f.path.match(/\.[^./\\]+$/)?.[0]?.toLowerCase() || '';
      if (nonCodeExts.has(ext)) return false;
      return true;
    });

    // Deterministic sort by path using < > (not localeCompare)
    relevant.sort((a, b) => (a.path < b.path ? -1 : a.path > b.path ? 1 : 0));

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
    const nonZero = relevant.filter((f) => f.confidence > 0);
    const confidenceAvg =
      nonZero.length > 0
        ? Math.round(
            (nonZero.reduce((sum, f) => sum + f.confidence, 0) /
              nonZero.length) *
              100
          ) / 100
        : 0;

    const testFileCount = relevant.filter((f) => f.type === 'test').length;

    return {
      schemaVersion: '1.0.0',
      meta: {
        hamletVersion,
        nodeVersion: process.version,
        generatedAt: new Date().toISOString(),
        root: resolvedRoot,
      },
      summary: {
        fileCount: relevant.length,
        testFileCount,
        frameworksDetected: detectedFrameworks,
        directionsSupported,
        confidenceAvg,
      },
      files: relevant,
    };
  }

  /**
   * Compute a meaningful confidence score by blending classification base,
   * framework detection quality, and content complexity.
   *
   * @param {{type: string, framework: string|null, confidence: number}} classification
   * @param {{framework: string, score: number}[]} candidates
   * @param {string} content - File content for complexity analysis
   * @returns {number} 0-100
   */
  _computeConfidence(classification, candidates, content) {
    const base = classification.confidence;
    if (base === 0) return 0;

    // Config and type-def files: classification is path-based, keep as-is
    if (
      classification.type === 'config' ||
      classification.type === 'type-def'
    ) {
      return base;
    }

    // --- Detection quality factor (0.4–1.0) ---
    let detectionFactor;
    if (candidates.length === 0) {
      detectionFactor = classification.framework ? 0.6 : 0.4;
    } else {
      const topScore = candidates[0].score;
      const secondScore = candidates.length > 1 ? candidates[1].score : 0;
      const discrimination =
        secondScore > 0 ? (topScore - secondScore) / topScore : 1;
      const magnitude = Math.min(topScore / 20, 1);
      detectionFactor = 0.4 + 0.3 * discrimination + 0.3 * magnitude;
    }

    // --- Complexity factor (0.55–1.0) ---
    // Simple tests → high factor (easy to convert)
    // Complex tests → lower factor (harder to convert reliably)
    const complexity = this._analyzeComplexity(content);
    const complexityFactor = 1.0 - complexity * 0.45;

    return Math.min(
      99,
      Math.max(10, Math.round(base * detectionFactor * complexityFactor))
    );
  }

  /**
   * Analyze test content complexity on a 0–1 scale.
   * 0 = trivial, 1 = maximally complex.
   *
   * @param {string} content
   * @returns {number} 0-1
   */
  _analyzeComplexity(content) {
    const lines = content.split('\n');
    const lineCount = lines.length;

    // Size contribution (0–0.25)
    let size = 0;
    if (lineCount > 500) size = 0.25;
    else if (lineCount > 200) size = 0.18;
    else if (lineCount > 100) size = 0.1;
    else if (lineCount > 50) size = 0.05;

    // Test count — many tests = more conversion surface
    const testMatches = content.match(
      /\b(?:it|test)\s*\(|@Test\b|def\s+test_/g
    );
    const testCount = testMatches ? testMatches.length : 0;
    let tests = 0;
    if (testCount > 30) tests = 0.15;
    else if (testCount > 15) tests = 0.1;
    else if (testCount > 5) tests = 0.05;

    // Nesting depth — deep describe/context nesting
    let maxDepth = 0;
    let depth = 0;
    for (const line of lines) {
      if (/\b(?:describe|context|suite)\s*\(/.test(line)) {
        depth++;
        if (depth > maxDepth) maxDepth = depth;
      }
      // Closing of describe-like blocks (heuristic: line with just });)
      if (/^\s*\}\s*\)\s*;?\s*$/.test(line) && depth > 0) {
        depth--;
      }
    }
    let nesting = 0;
    if (maxDepth >= 4) nesting = 0.15;
    else if (maxDepth >= 3) nesting = 0.1;
    else if (maxDepth >= 2) nesting = 0.05;

    // Mocking — mock/spy patterns are hard to convert between frameworks
    const mockPatterns =
      /\bjest\.mock\b|\bjest\.fn\b|\bjest\.spyOn\b|\bvi\.mock\b|\bvi\.fn\b|\bvi\.spyOn\b|\bsinon\.\w+|\bcreateSpyObj\b|\bcreate_autospec\b|\b@mock\b|\bMockBean\b/gi;
    const mockMatches = content.match(mockPatterns);
    const mockCount = mockMatches ? mockMatches.length : 0;
    let mocking = 0;
    if (mockCount > 10) mocking = 0.2;
    else if (mockCount > 3) mocking = 0.12;
    else if (mockCount > 0) mocking = 0.06;

    // Async complexity
    const asyncPatterns =
      /\basync\b|\bawait\b|\.then\s*\(|\bsetTimeout\b|\buseFakeTimers\b|\bfakeAsync\b|\badvanceTimersByTime\b/g;
    const asyncMatches = content.match(asyncPatterns);
    const asyncCount = asyncMatches ? asyncMatches.length : 0;
    let async = 0;
    if (asyncCount > 15) async = 0.12;
    else if (asyncCount > 5) async = 0.06;
    else if (asyncCount > 0) async = 0.02;

    // Hooks — setup/teardown add conversion complexity
    const hookPatterns =
      /\b(?:beforeEach|afterEach|beforeAll|afterAll|before|after|setUp|tearDown|@Before|@After|@BeforeEach|@AfterEach)\b/g;
    const hookMatches = content.match(hookPatterns);
    const hookCount = hookMatches ? hookMatches.length : 0;
    let hooks = 0;
    if (hookCount > 6) hooks = 0.08;
    else if (hookCount > 2) hooks = 0.04;
    else if (hookCount > 0) hooks = 0.02;

    // Parameterized tests
    const paramPatterns =
      /\b(?:test\.each|it\.each|describe\.each|@ParameterizedTest|@DataProvider|@pytest\.mark\.parametrize)\b/g;
    const paramMatches = content.match(paramPatterns);
    let params = paramMatches ? 0.06 : 0;

    const total = size + tests + nesting + mocking + async + hooks + params;
    return Math.min(1, total);
  }
}
