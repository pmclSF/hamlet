/**
 * Master orchestrator for project-wide migration.
 *
 * Pipeline: scan → classify → build graph → sort → convert → rewrite imports → generate checklist
 *
 * Supports resume (--continue) and retry-failed (--retry-failed).
 */

import fs from 'fs/promises';
import path from 'path';
import { Scanner } from './Scanner.js';
import { FileClassifier } from './FileClassifier.js';
import { DependencyGraphBuilder } from './DependencyGraphBuilder.js';
import { TopologicalSorter } from './TopologicalSorter.js';
import { InputNormalizer } from './InputNormalizer.js';
import { ErrorRecovery } from './ErrorRecovery.js';
import { OutputValidator } from './OutputValidator.js';
import { ImportRewriter } from './ImportRewriter.js';
import { MigrationStateManager } from './MigrationStateManager.js';
import { MigrationChecklistGenerator } from './MigrationChecklistGenerator.js';
import { ConverterFactory } from './ConverterFactory.js';

export class MigrationEngine {
  constructor() {
    this.scanner = new Scanner();
    this.classifier = new FileClassifier();
    this.graphBuilder = new DependencyGraphBuilder();
    this.sorter = new TopologicalSorter();
    this.normalizer = new InputNormalizer();
    this.errorRecovery = new ErrorRecovery();
    this.validator = new OutputValidator();
    this.importRewriter = new ImportRewriter();
    this.checklistGenerator = new MigrationChecklistGenerator();
  }

  /**
   * Run a full project migration.
   *
   * @param {string} rootDir - Project root directory
   * @param {Object} options
   * @param {string} options.from - Source framework
   * @param {string} options.to - Target framework
   * @param {string} [options.output] - Output directory (if different from source)
   * @param {boolean} [options.continue] - Resume from previous state
   * @param {boolean} [options.retryFailed] - Retry only failed files
   * @param {Function} [options.onProgress] - Progress callback(file, status, confidence)
   * @returns {Promise<{results: Array, checklist: string, state: Object}>}
   */
  async migrate(rootDir, options) {
    const { from, to } = options;
    const resolvedRoot = path.resolve(rootDir);

    // State management
    const stateManager = new MigrationStateManager(resolvedRoot);
    let isResume = false;

    if (options.continue || options.retryFailed) {
      try {
        await stateManager.load();
        isResume = true;
      } catch {
        // No existing state — start fresh
        await stateManager.init({ source: from, target: to });
      }
    } else {
      await stateManager.init({ source: from, target: to });
    }

    // 1. Scan
    const scanned = await this.scanner.scan(resolvedRoot, {
      include: ['*.js', '*.ts', '*.jsx', '*.tsx', '*.mjs'],
    });

    // 2. Read contents and classify
    const files = [];
    for (const entry of scanned) {
      let content;
      try {
        content = await fs.readFile(entry.path, 'utf8');
      } catch {
        continue;
      }

      const classification = this.classifier.classify(entry.relativePath, content);
      files.push({
        ...entry,
        content,
        classification,
      });
    }

    // 3. Build dependency graph
    const graph = this.graphBuilder.build(files);

    // 4. Sort topologically
    const sortedPaths = this.sorter.sort(graph);

    // 5. Create converter
    let converter;
    try {
      converter = await ConverterFactory.createConverter(from, to);
    } catch (error) {
      throw new Error(`Failed to create converter for ${from}→${to}: ${error.message}`);
    }

    // 6. Convert files in order
    const results = [];
    const renames = new Map();

    for (const filePath of sortedPaths) {
      const file = files.find(f => f.path === filePath);
      if (!file) continue;

      // Skip non-convertible types
      if (file.classification.type === 'fixture' || file.classification.type === 'type-def') {
        stateManager.markFileSkipped(file.relativePath, `Non-convertible type: ${file.classification.type}`);
        if (options.onProgress) options.onProgress(file.relativePath, 'skipped', 0);
        continue;
      }

      // Resume logic
      if (isResume && !options.retryFailed && stateManager.isConverted(file.relativePath)) {
        if (options.onProgress) options.onProgress(file.relativePath, 'skipped-converted', null);
        continue;
      }

      if (isResume && options.retryFailed && !stateManager.isFailed(file.relativePath)) {
        if (options.onProgress) options.onProgress(file.relativePath, 'skipped', null);
        continue;
      }

      // Normalize input
      const { normalized, issues: normIssues } = this.normalizer.normalize(file.content);

      if (normIssues.some(i => i.type === 'binary')) {
        stateManager.markFileSkipped(file.relativePath, 'Binary file');
        results.push({
          path: file.relativePath,
          confidence: 0,
          status: 'skipped',
          warnings: ['Binary file detected'],
          todos: [],
          type: file.classification.type,
        });
        if (options.onProgress) options.onProgress(file.relativePath, 'skipped', 0);
        continue;
      }

      // Convert
      let converted;
      let confidence = 0;
      const warnings = normIssues.map(i => i.message);
      const todos = [];

      try {
        const convResult = await converter.convert(normalized);
        converted = convResult;

        // Get confidence from last report if available
        if (converter.getLastReport) {
          const report = converter.getLastReport();
          confidence = report ? report.confidence : 85;
        } else {
          confidence = 85;
        }
      } catch (convError) {
        // Try error recovery
        try {
          const { recovered, warnings: recWarnings } = this.errorRecovery.recoverFromParseError(
            normalized,
            convError,
            (line) => line
          );
          converted = recovered;
          warnings.push(...recWarnings);
          confidence = 30;
        } catch {
          stateManager.markFileConverted(file.relativePath, { confidence: 0, error: convError.message });
          results.push({
            path: file.relativePath,
            confidence: 0,
            status: 'failed',
            error: convError.message,
            warnings,
            todos: [],
            type: file.classification.type,
          });
          if (options.onProgress) options.onProgress(file.relativePath, 'failed', 0);
          continue;
        }
      }

      // Validate output
      const validation = this.validator.validate(converted, to);
      if (!validation.valid) {
        for (const issue of validation.issues) {
          warnings.push(issue.message);
        }
        // Lower confidence for validation issues
        confidence = Math.min(confidence, 70);
      }

      // Track renames for import rewriting
      if (file.relativePath !== this._computeNewPath(file.relativePath, from, to)) {
        renames.set(
          './' + file.relativePath.replace(/\\/g, '/'),
          './' + this._computeNewPath(file.relativePath, from, to).replace(/\\/g, '/')
        );
      }

      // Write output
      const outputDir = options.output ? path.resolve(options.output) : resolvedRoot;
      const newRelPath = this._computeNewPath(file.relativePath, from, to);
      const outputPath = path.join(outputDir, newRelPath);
      await fs.mkdir(path.dirname(outputPath), { recursive: true });
      await fs.writeFile(outputPath, converted, 'utf8');

      stateManager.markFileConverted(file.relativePath, { confidence });
      results.push({
        path: file.relativePath,
        confidence,
        status: 'converted',
        warnings,
        todos,
        type: file.classification.type,
      });
      if (options.onProgress) options.onProgress(file.relativePath, 'converted', confidence);
    }

    // 7. Rewrite imports in converted files
    if (renames.size > 0 && options.output) {
      const outputDir = path.resolve(options.output);
      for (const result of results) {
        if (result.status !== 'converted') continue;
        const newRelPath = this._computeNewPath(result.path, from, to);
        const filePath = path.join(outputDir, newRelPath);
        try {
          const content = await fs.readFile(filePath, 'utf8');
          const rewritten = this.importRewriter.rewrite(content, renames);
          if (rewritten !== content) {
            await fs.writeFile(filePath, rewritten, 'utf8');
          }
        } catch {
          // File may not exist if conversion failed
        }
      }
    }

    // 8. Save state
    await stateManager.save();

    // 9. Generate checklist
    const checklist = this.checklistGenerator.generate(graph, results);

    return { results, checklist, state: stateManager.getStatus() };
  }

  /**
   * Compute new file path based on framework conventions.
   *
   * @param {string} relativePath
   * @param {string} from
   * @param {string} to
   * @returns {string}
   */
  _computeNewPath(relativePath, from, to) {
    let newPath = relativePath;

    // Extension-based renames
    if (from === 'jest' && to === 'vitest') {
      // .test.js stays .test.js for vitest
      // no change needed
    } else if (from === 'cypress' && to === 'playwright') {
      newPath = newPath.replace(/\.cy\.(js|ts|jsx|tsx)$/, '.spec.$1');
    } else if (from === 'playwright' && to === 'cypress') {
      newPath = newPath.replace(/\.spec\.(js|ts|jsx|tsx)$/, '.cy.$1');
    }

    return newPath;
  }
}
