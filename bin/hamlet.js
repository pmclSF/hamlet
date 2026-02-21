#!/usr/bin/env node

import { program } from 'commander';
import chalk from 'chalk';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { createRequire } from 'module';
import fs from 'fs/promises';
import path from 'path';
import fg from 'fast-glob';
import {
  convertFile,
  convertRepository,
  validateTests,
  generateReport,
  TestValidator,
  ConversionReporter,
} from '../src/index.js';
import { ConverterFactory, FRAMEWORKS } from '../src/core/ConverterFactory.js';
import { FrameworkDetector } from '../src/core/FrameworkDetector.js';
import {
  SHORTHANDS,
  CONVERSION_CATEGORIES,
  FRAMEWORK_ABBREV,
} from '../src/cli/shorthands.js';

const require = createRequire(import.meta.url);
const version = require('../package.json').version;

// ── Color / TTY detection ────────────────────────────────────────────
const isTTY = process.stdout.isTTY;
const useColor =
  isTTY && !process.env.NO_COLOR && !process.argv.includes('--no-color');
if (!useColor) {
  chalk.level = 0;
}

// ── Framework language map (for cross-language error messages) ────────
const FRAMEWORK_LANGUAGE = {
  cypress: 'javascript',
  playwright: 'javascript',
  selenium: 'javascript',
  jest: 'javascript',
  vitest: 'javascript',
  mocha: 'javascript',
  jasmine: 'javascript',
  junit4: 'java',
  junit5: 'java',
  testng: 'java',
  pytest: 'python',
  unittest: 'python',
  nose2: 'python',
  webdriverio: 'javascript',
  puppeteer: 'javascript',
  testcafe: 'javascript',
};

// ── Output filename helpers ──────────────────────────────────────────
function getTargetExtension(toFramework, originalExt) {
  if (toFramework === 'cypress') return '.cy' + (originalExt || '.js');
  if (toFramework === 'playwright') return '.spec' + (originalExt || '.js');
  return '.test' + (originalExt || '.js');
}

function buildOutputFilename(sourceBasename, toFramework) {
  const ext = path.extname(sourceBasename);
  const base = path.basename(sourceBasename, ext);
  const cleanBase = base.replace(/\.(cy|spec|test)$/, '');
  if (toFramework === 'cypress') return cleanBase + '.cy.js';
  if (toFramework === 'playwright') return cleanBase + '.spec.js';
  return cleanBase + '.test.js';
}

// ── Progress display ─────────────────────────────────────────────────
function showProgress(current, total, currentFile) {
  if (!isTTY) return;
  const pct = Math.round((current / total) * 100);
  const filled = Math.round((current / total) * 20);
  const bar = '\u2588'.repeat(filled) + '\u2591'.repeat(20 - filled);
  process.stdout.write(
    `\r  [${bar}] ${current}/${total} (${pct}%) ${currentFile}`
  );
}

function clearProgress() {
  if (!isTTY) return;
  process.stdout.write('\r' + ' '.repeat(80) + '\r');
}

function shouldShowStack() {
  return program.opts().debug || !!process.env.DEBUG;
}

// ── Extracted convertAction ──────────────────────────────────────────
async function convertAction(source, options) {
  const quiet = options.quiet || false;
  const verbose = options.verbose || program.opts().verbose || false;
  const jsonOutput = options.json || false;
  const plan = options.plan || false;
  const dryRun = options.dryRun || plan;
  const onError = options.onError || 'skip';

  // Validate frameworks
  const validFrameworks = Object.values(FRAMEWORKS);
  const fromFramework = options.from.toLowerCase();
  const toFramework = options.to.toLowerCase();

  if (!validFrameworks.includes(fromFramework)) {
    // Check for cross-language hint
    const msg = `Invalid source framework: ${options.from}. Valid options: ${validFrameworks.join(', ')}`;
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: msg }));
    } else {
      console.error(chalk.red('Error:'), msg);
      console.error(
        chalk.gray('Next steps: Run `hamlet list` to see supported frameworks.')
      );
    }
    process.exit(2);
  }

  if (!validFrameworks.includes(toFramework)) {
    const msg = `Invalid target framework: ${options.to}. Valid options: ${validFrameworks.join(', ')}`;
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: msg }));
    } else {
      console.error(chalk.red('Error:'), msg);
      console.error(
        chalk.gray('Next steps: Run `hamlet list` to see supported frameworks.')
      );
    }
    process.exit(2);
  }

  if (fromFramework === toFramework) {
    const msg = 'Source and target frameworks must be different';
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: msg }));
    } else {
      console.error(chalk.red('Error:'), msg);
      console.error(
        chalk.gray('Next steps: Specify different --from and --to frameworks.')
      );
    }
    process.exit(2);
  }

  // Cross-language check
  const fromLang = FRAMEWORK_LANGUAGE[fromFramework];
  const toLang = FRAMEWORK_LANGUAGE[toFramework];
  if (fromLang && toLang && fromLang !== toLang) {
    const msg = `Cannot convert between ${fromFramework} (${fromLang}) and ${toFramework} (${toLang}). Hamlet only converts within the same language.`;
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: msg }));
    } else {
      console.error(chalk.red('Error:'), msg);
      console.error(
        chalk.gray(
          'Next steps: Hamlet converts within the same language. See `hamlet list` for valid directions.'
        )
      );
    }
    process.exit(2);
  }

  // Check if direction is supported
  if (!ConverterFactory.isSupported(fromFramework, toFramework)) {
    const supported = ConverterFactory.getSupportedConversions()
      .filter((c) => c.startsWith(fromFramework + '-'))
      .map((c) => c.split('-')[1]);
    let msg = `Unsupported conversion: ${fromFramework} to ${toFramework}.`;
    if (supported.length > 0) {
      msg += ` Supported targets for ${fromFramework}: ${supported.join(', ')}`;
    }
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: msg }));
    } else {
      console.error(chalk.red('Error:'), msg);
      console.error(
        chalk.gray('Next steps: Run `hamlet list` to see supported frameworks.')
      );
    }
    process.exit(2);
  }

  if (!quiet && !jsonOutput) {
    console.log(
      chalk.blue(
        `Converting from ${chalk.bold(fromFramework)} to ${chalk.bold(toFramework)}...`
      )
    );
  }

  // Auto-detect framework if requested
  if (options.autoDetect) {
    try {
      const content = await fs.readFile(source, 'utf8');
      const detection = FrameworkDetector.detectFromContent(content);
      if (detection.framework && detection.confidence > 0.5) {
        if (!quiet && !jsonOutput) {
          console.log(
            chalk.yellow(
              `Auto-detected source framework: ${detection.framework} (${Math.round(detection.confidence * 100)}% confidence)`
            )
          );
        }
        options.from = detection.framework;
      }
    } catch (_e) {
      // Ignore auto-detection errors
    }
  }

  // Determine if source is a file, directory, or glob
  let sourceFiles = [];
  let isBatch = false;
  let sourceRoot = '';

  // Check for glob characters
  const isGlob =
    source.includes('*') || source.includes('?') || source.includes('{');

  if (isGlob) {
    // Glob pattern
    const matches = await fg(source, { absolute: true });
    if (matches.length === 0) {
      if (jsonOutput) {
        console.log(
          JSON.stringify({
            success: true,
            files: [],
            summary: { converted: 0, skipped: 0, failed: 0 },
          })
        );
      } else if (!quiet) {
        console.log(chalk.yellow(`No files matched pattern: ${source}`));
      }
      return;
    }
    sourceFiles = matches.map((f) => ({
      path: f,
      relativePath: path.basename(f),
    }));
    sourceRoot = path.dirname(matches[0]);
    isBatch = true;
  } else {
    // Check if file or directory
    let sourceStat;
    try {
      sourceStat = await fs.stat(source);
    } catch (_e) {
      // File not found — try to suggest similar files
      const msg = `File not found: ${source}`;
      try {
        const parentDir = path.dirname(source);
        const basename = path.basename(source);
        const entries = await fs.readdir(parentDir);
        const similar = entries
          .filter((e) => {
            const lower = e.toLowerCase();
            const targetLower = basename.toLowerCase();
            return (
              lower.includes(targetLower.slice(0, 4)) ||
              targetLower.includes(lower.slice(0, 4))
            );
          })
          .slice(0, 3);
        if (similar.length > 0) {
          const suggestion = `\nDid you mean: ${similar.join(', ')}?`;
          if (jsonOutput) {
            console.log(
              JSON.stringify({ success: false, error: msg + suggestion })
            );
          } else {
            console.error(chalk.red('Error:'), msg);
            console.error(chalk.yellow(suggestion));
          }
        } else {
          if (jsonOutput) {
            console.log(JSON.stringify({ success: false, error: msg }));
          } else {
            console.error(chalk.red('Error:'), msg);
            console.error(
              chalk.gray(
                'Next steps: Check the file path and ensure the file exists.'
              )
            );
          }
        }
      } catch (_readErr) {
        if (jsonOutput) {
          console.log(JSON.stringify({ success: false, error: msg }));
        } else {
          console.error(chalk.red('Error:'), msg);
          console.error(
            chalk.gray(
              'Next steps: Check the file path and ensure the file exists.'
            )
          );
        }
      }
      process.exit(2);
    }

    if (sourceStat.isDirectory()) {
      // Directory mode — require --output
      if (!options.output && !dryRun) {
        const msg = '--output is required when converting a directory';
        if (jsonOutput) {
          console.log(JSON.stringify({ success: false, error: msg }));
        } else {
          console.error(chalk.red('Error:'), msg);
          console.error(
            chalk.gray(
              'Next steps: Specify an output directory with -o <path>.'
            )
          );
        }
        process.exit(2);
      }

      // Use Scanner + FileClassifier for better file discovery
      const { Scanner } = await import('../src/core/Scanner.js');
      const { FileClassifier } = await import('../src/core/FileClassifier.js');
      const scanner = new Scanner();
      const classifier = new FileClassifier();

      const allFiles = await scanner.scan(source);
      sourceRoot = path.resolve(source);

      for (const file of allFiles) {
        try {
          const content = await fs.readFile(file.path, 'utf8');
          const classification = classifier.classify(file.path, content);
          if (
            classification.type === 'test' &&
            classification.framework === fromFramework
          ) {
            sourceFiles.push(file);
          }
        } catch (_e) {
          // Skip unreadable files
        }
      }

      if (sourceFiles.length === 0) {
        // Fallback: try all JS/TS files if classifier found nothing
        const fallbackFiles = allFiles.filter((f) =>
          /\.(js|ts|tsx|jsx|py|java|rb)$/.test(f.path)
        );
        if (fallbackFiles.length > 0) {
          sourceFiles = fallbackFiles;
        }
      }

      isBatch = true;
    } else {
      sourceFiles = [
        { path: path.resolve(source), relativePath: path.basename(source) },
      ];
    }
  }

  // Batch mode with --output required for glob too
  if (isBatch && !options.output && !dryRun) {
    const msg = '--output is required when converting multiple files';
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: msg }));
    } else {
      console.error(chalk.red('Error:'), msg);
      console.error(
        chalk.gray('Next steps: Specify an output directory with -o <path>.')
      );
    }
    process.exit(2);
  }

  // Create converter
  let converter;
  try {
    converter = await ConverterFactory.createConverter(
      fromFramework,
      toFramework,
      {
        batchSize: parseInt(options.batchSize || '5'),
        preserveStructure: options.preserveStructure,
      }
    );
  } catch (error) {
    if (jsonOutput) {
      console.log(JSON.stringify({ success: false, error: error.message }));
    } else {
      console.error(chalk.red('Error:'), error.message);
    }
    if (shouldShowStack()) console.error(error.stack);
    process.exit(1);
  }

  // ── Dry-run / Plan mode ──────────────────────────────────────────
  if (dryRun) {
    if (isBatch) {
      // Batch dry-run: compute per-file confidence
      const counts = { high: 0, medium: 0, low: 0 };
      const fileDetails = [];
      const outputDir = options.output ? path.resolve(options.output) : null;

      for (const file of sourceFiles) {
        const relPath = file.relativePath || path.basename(file.path);
        const newFilename = buildOutputFilename(
          path.basename(file.path),
          toFramework
        );
        const relDir = path.dirname(relPath);
        const outputFilePath = outputDir
          ? path.join(outputDir, relDir === '.' ? '' : relDir, newFilename)
          : newFilename;

        let conf = 0;
        let level = 'low';
        try {
          const content = await fs.readFile(file.path, 'utf8');
          await converter.convert(content);
          const report = converter.getLastReport
            ? converter.getLastReport()
            : null;
          if (report) {
            conf = report.confidence || 0;
          } else {
            conf = 95;
          }
        } catch (_e) {
          conf = 0;
        }

        if (conf >= 80) {
          level = 'high';
          counts.high++;
        } else if (conf >= 50) {
          level = 'medium';
          counts.medium++;
        } else {
          counts.low++;
        }

        fileDetails.push({
          source: relPath,
          sourceFull: file.path,
          output: outputFilePath,
          confidence: conf,
          level,
        });
      }

      if (plan) {
        // ── Plan output (batch) ────────────────────────────────────
        if (jsonOutput) {
          console.log(
            JSON.stringify({
              plan: true,
              direction: { from: fromFramework, to: toFramework },
              files: fileDetails.map((f) => ({
                source: f.source,
                output: f.output,
                confidence: f.confidence,
              })),
              summary: {
                total: fileDetails.length,
                confidence: counts,
              },
              warnings: [],
            })
          );
        } else if (!quiet) {
          console.log(
            chalk.bold(
              `Conversion Plan: ${fromFramework} \u2192 ${toFramework}`
            )
          );
          console.log(`  ${fileDetails.length} files to convert\n`);
          console.log(`  ${'Input'.padEnd(30)}    ${'Output'}`);
          for (const f of fileDetails) {
            console.log(`  ${f.source.padEnd(30)} \u2192  ${f.output}`);
          }
          console.log(
            `\n  Confidence: ${counts.high} high, ${counts.medium} medium, ${counts.low} low`
          );
          console.log('  Warnings: (none)');
        }
      } else {
        // ── Standard dry-run output (batch) ────────────────────────
        if (jsonOutput) {
          console.log(
            JSON.stringify({
              success: true,
              dryRun: true,
              files: sourceFiles.map((f) => ({
                source: f.relativePath || f.path,
              })),
              summary: {
                converted: sourceFiles.length,
                skipped: 0,
                failed: 0,
                confidence: counts,
              },
            })
          );
        } else if (!quiet) {
          console.log(
            chalk.yellow('Dry run mode - no files will be modified\n')
          );
          console.log(`  Files found: ${sourceFiles.length}`);
          console.log(`  Would convert: ${sourceFiles.length}`);
          console.log(`\n  Confidence distribution:`);
          console.log(`    ${chalk.green('High:')}   ${counts.high}`);
          console.log(`    ${chalk.yellow('Medium:')} ${counts.medium}`);
          console.log(`    ${chalk.red('Low:')}    ${counts.low}`);
        }
      }
    } else {
      // Single file dry-run / plan
      const filePath = sourceFiles[0].path;
      const relSource = sourceFiles[0].relativePath || path.basename(filePath);

      // Compute output path
      let outputPath = options.output;
      if (outputPath) {
        if (!path.extname(outputPath)) {
          outputPath = path.join(
            outputPath,
            buildOutputFilename(path.basename(source), toFramework)
          );
        }
      } else {
        const ext = path.extname(source);
        const base = path.basename(source, ext);
        const dir = path.dirname(source);
        const newExt = getTargetExtension(toFramework, ext);
        outputPath = path.join(
          dir,
          base.replace(/\.(cy|spec|test)$/, '') + newExt
        );
      }

      try {
        const content = await fs.readFile(filePath, 'utf8');
        await converter.convert(content);
        const report = converter.getLastReport
          ? converter.getLastReport()
          : null;
        const conf = report ? report.confidence || 0 : 0;
        const level =
          (report && report.level) ||
          (conf >= 80 ? 'high' : conf >= 50 ? 'medium' : 'low');

        if (plan) {
          // ── Plan output (single) ──────────────────────────────
          if (jsonOutput) {
            console.log(
              JSON.stringify({
                plan: true,
                direction: { from: fromFramework, to: toFramework },
                files: [
                  {
                    source: relSource,
                    output: outputPath,
                    confidence: conf,
                  },
                ],
                summary: {
                  total: 1,
                  confidence: {
                    high: level === 'high' ? 1 : 0,
                    medium: level === 'medium' ? 1 : 0,
                    low: level === 'low' ? 1 : 0,
                  },
                },
                warnings: [],
              })
            );
          } else if (!quiet) {
            console.log(chalk.bold('Conversion Plan:'));
            console.log(`  ${relSource} \u2192 ${outputPath}`);
            console.log(`  Direction: ${fromFramework} \u2192 ${toFramework}`);
            console.log(`  Confidence: ${conf}% (${level})`);
          }
        } else {
          // ── Standard dry-run output (single) ──────────────────
          if (jsonOutput) {
            console.log(
              JSON.stringify({
                success: true,
                dryRun: true,
                files: [
                  {
                    source: filePath,
                    confidence: report ? report.confidence : null,
                  },
                ],
                summary: {
                  converted: 1,
                  skipped: 0,
                  failed: 0,
                },
              })
            );
          } else if (!quiet) {
            console.log(
              chalk.yellow('Dry run mode - no files will be modified\n')
            );
            console.log(`  Would convert: ${filePath}`);
            console.log(
              `  Output: ${options.output || 'same directory with new extension'}`
            );
            if (report) {
              console.log(`  Confidence: ${conf}% (${level})`);
              if (verbose) {
                console.log(`\n  Details:`);
                if (report.converted != null)
                  console.log(`    ${report.converted} patterns converted`);
                if (report.warnings != null)
                  console.log(`    ${report.warnings} warnings`);
                if (report.unconvertible != null)
                  console.log(`    ${report.unconvertible} unconvertible`);
              }
            }
          }
        }
      } catch (error) {
        if (jsonOutput) {
          console.log(
            JSON.stringify({
              success: false,
              dryRun: true,
              files: [{ source: filePath, error: error.message }],
              summary: { converted: 0, skipped: 0, failed: 1 },
            })
          );
        } else {
          console.error(chalk.red('Error:'), error.message);
        }
        process.exit(1);
      }
    }
    return;
  }

  // ── Actual conversion ────────────────────────────────────────────
  const isRepository =
    !isBatch &&
    (source.includes('github.com') || source.includes('gitlab.com'));

  if (isRepository) {
    await convertRepository(source, options.output, {
      ...options,
      converter,
      fromFramework,
      toFramework,
    });
    if (!quiet && !jsonOutput) {
      console.log(chalk.green('\nConversion completed successfully!'));
    }
    return;
  }

  if (isBatch) {
    // ── Batch conversion ───────────────────────────────────────────
    const results = { converted: 0, skipped: 0, failed: 0, files: [] };
    const total = sourceFiles.length;

    if (total === 0) {
      if (jsonOutput) {
        console.log(
          JSON.stringify({ success: true, files: [], summary: results })
        );
      } else if (!quiet) {
        console.log(chalk.yellow('No matching files found to convert.'));
      }
      return;
    }

    const outputDir = path.resolve(options.output);
    await fs.mkdir(outputDir, { recursive: true });

    for (let i = 0; i < total; i++) {
      const file = sourceFiles[i];
      const relPath = file.relativePath || path.basename(file.path);
      const newFilename = buildOutputFilename(
        path.basename(file.path),
        toFramework
      );
      const relDir = path.dirname(relPath);
      const outputFilePath = path.join(
        outputDir,
        relDir === '.' ? '' : relDir,
        newFilename
      );

      if (!quiet && !jsonOutput && isTTY) {
        showProgress(i + 1, total, path.basename(file.path));
      }

      try {
        const content = await fs.readFile(file.path, 'utf8');
        const converted = await converter.convert(content, options);
        const report = converter.getLastReport
          ? converter.getLastReport()
          : null;

        await fs.mkdir(path.dirname(outputFilePath), { recursive: true });
        await fs.writeFile(outputFilePath, converted);

        results.converted++;
        results.files.push({
          source: file.path,
          output: outputFilePath,
          confidence: report ? report.confidence : null,
        });

        if (!quiet && !jsonOutput && !isTTY) {
          console.log(
            chalk.green(
              `  \u2713 ${relPath} -> ${path.relative(outputDir, outputFilePath)}`
            )
          );
        }
        if (verbose && !jsonOutput) {
          if (report) {
            console.log(chalk.gray(`    Confidence: ${report.confidence}%`));
          }
        }
      } catch (error) {
        if (onError === 'fail') {
          if (!quiet && !jsonOutput) {
            clearProgress();
            console.error(chalk.red(`\n  \u2717 ${relPath}: ${error.message}`));
          }
          results.failed++;
          results.files.push({ source: file.path, error: error.message });
          if (jsonOutput) {
            console.log(
              JSON.stringify({
                success: false,
                files: results.files,
                summary: results,
              })
            );
          }
          process.exit(1);
        } else if (onError === 'best-effort') {
          // Write partial output with warning comment
          const partialContent = `// HAMLET-WARNING: Conversion incomplete - ${error.message}\n`;
          try {
            await fs.mkdir(path.dirname(outputFilePath), { recursive: true });
            await fs.writeFile(outputFilePath, partialContent);
          } catch (_writeErr) {
            // Ignore write errors for partial content
          }
          results.failed++;
          results.files.push({
            source: file.path,
            error: error.message,
            partial: true,
          });
          if (!quiet && !jsonOutput) {
            if (isTTY) clearProgress();
            console.log(
              chalk.yellow(
                `  ! ${relPath}: ${error.message} (partial output written)`
              )
            );
          }
        } else {
          // skip mode (default)
          results.skipped++;
          results.files.push({
            source: file.path,
            error: error.message,
            skipped: true,
          });
          if (!quiet && !jsonOutput) {
            if (isTTY) clearProgress();
            console.log(
              chalk.yellow(`  - ${relPath}: skipped (${error.message})`)
            );
          }
        }
      }
    }

    if (isTTY && !quiet && !jsonOutput) {
      clearProgress();
    }

    // Summary
    if (jsonOutput) {
      console.log(
        JSON.stringify({
          success: results.failed === 0,
          files: results.files,
          summary: results,
        })
      );
    } else if (!quiet) {
      console.log(
        chalk.bold(
          `\nSummary: ${chalk.green(results.converted + ' converted')}, ${chalk.yellow(results.skipped + ' skipped')}, ${chalk.red(results.failed + ' failed')}`
        )
      );
    }

    if (results.failed > 0) {
      process.exit(results.converted > 0 ? 3 : 1);
    }
  } else {
    // ── Single file conversion ─────────────────────────────────────
    const filePath = sourceFiles[0].path;

    // Determine output path
    let outputPath = options.output;

    if (outputPath) {
      try {
        const outputStat = await fs.stat(outputPath);
        if (outputStat.isDirectory()) {
          outputPath = path.join(
            outputPath,
            buildOutputFilename(path.basename(source), toFramework)
          );
        }
      } catch (_e) {
        // Output path doesn't exist yet
        if (!path.extname(outputPath)) {
          // No extension — treat as directory
          await fs.mkdir(outputPath, { recursive: true });
          outputPath = path.join(
            outputPath,
            buildOutputFilename(path.basename(source), toFramework)
          );
        }
      }
    }

    if (!outputPath) {
      const ext = path.extname(source);
      const base = path.basename(source, ext);
      const dir = path.dirname(source);
      const newExt = getTargetExtension(toFramework, ext);
      outputPath = path.join(
        dir,
        base.replace(/\.(cy|spec|test)$/, '') + newExt
      );
    }

    // Read source file
    const content = await fs.readFile(filePath, 'utf8');

    // Convert
    let converted;
    try {
      converted = await converter.convert(content, options);
    } catch (error) {
      if (onError === 'best-effort') {
        converted = `// HAMLET-WARNING: Conversion incomplete - ${error.message}\n`;
        await fs.mkdir(path.dirname(outputPath), { recursive: true });
        await fs.writeFile(outputPath, converted);
        if (jsonOutput) {
          console.log(
            JSON.stringify({
              success: false,
              files: [
                { source: filePath, error: error.message, partial: true },
              ],
              summary: { converted: 0, skipped: 0, failed: 1 },
            })
          );
        } else if (!quiet) {
          console.log(chalk.yellow(`Partial output written: ${outputPath}`));
        }
        process.exit(1);
      }
      throw error;
    }

    const report = converter.getLastReport ? converter.getLastReport() : null;

    // Write output
    await fs.mkdir(path.dirname(outputPath), { recursive: true });
    await fs.writeFile(outputPath, converted);

    if (jsonOutput) {
      const result = {
        success: true,
        files: [
          {
            source: filePath,
            output: outputPath,
            confidence: report ? report.confidence : null,
          },
        ],
        summary: { converted: 1, skipped: 0, failed: 0 },
      };
      console.log(JSON.stringify(result));
    } else if (!quiet) {
      console.log(
        chalk.green(`\n  \u2713 Converted: ${path.basename(source)}`)
      );
      if (report) {
        const conf = report.confidence || 0;
        const level =
          report.level || (conf >= 80 ? 'high' : conf >= 50 ? 'medium' : 'low');
        console.log(`  Confidence: ${conf}% (${level})`);
      }
      console.log(`  \u2192 Output: ${outputPath}`);
      if (verbose && report) {
        console.log(`\n  Details:`);
        if (report.converted != null)
          console.log(`    ${report.converted} patterns converted`);
        if (report.warnings != null)
          console.log(`    ${report.warnings} warnings`);
        if (report.unconvertible != null)
          console.log(`    ${report.unconvertible} unconvertible`);
      }
    }
  }

  if (options.validate && options.output) {
    if (!quiet && !jsonOutput) {
      console.log(chalk.blue('\nValidating converted tests...'));
    }
    await validateTests(options.output);
  }

  if (options.report && options.output) {
    if (!quiet && !jsonOutput) {
      console.log(chalk.blue('\nGenerating report...'));
    }
    await generateReport(options.output, options.report);
  }

  if (!isBatch && !jsonOutput && !quiet) {
    console.log(chalk.green('\nConversion completed successfully!'));
  }
}

// ── Program setup ────────────────────────────────────────────────────
program
  .version(version)
  .description(
    'Hamlet: Multi-framework test converter — 25 directions across JavaScript, Java, and Python.'
  )
  .addHelpText(
    'after',
    `
Shorthands:
  50 shorthand aliases are available (e.g. cy2pw, jest2vt, pyt2ut).
  Run ${chalk.cyan('hamlet shorthands')} to see them all.

Examples:
  $ hamlet convert src/tests/ --from jest --to vitest -o converted/
  $ hamlet jest2vt auth.test.js -o converted/
  $ hamlet migrate tests/ --from jest --to vitest -o out/
  $ hamlet estimate tests/ --from mocha --to jest
  $ hamlet detect src/auth.test.js
  $ hamlet doctor`
  )
  .option('--verbose', 'Show detailed output for all commands')
  .option('--debug', 'Show stack traces and internal debug info');

// ── Main convert command ─────────────────────────────────────────────
program
  .command('convert')
  .description('Convert tests between frameworks')
  .argument('<source>', 'Source test file, directory, or repository URL')
  .option('-f, --from <framework>', 'Source framework', 'cypress')
  .option('-t, --to <framework>', 'Target framework', 'playwright')
  .option('-o, --output <path>', 'Output path for converted tests')
  .option('-c, --config <path>', 'Custom configuration file path')
  .option('--test-type <type>', 'Specify test type (e2e, component, api, etc.)')
  .option('--validate', 'Validate converted tests')
  .option(
    '--report <format>',
    'Generate conversion report (html, json, markdown)'
  )
  .option('--preserve-structure', 'Maintain original directory structure')
  .option('--batch-size <number>', 'Number of files per batch', '5')
  .option('--dry-run', 'Show what would be converted without making changes')
  .option('--plan', 'Show structured conversion plan')
  .option('--auto-detect', 'Auto-detect source framework from file content')
  .option('-q, --quiet', 'Suppress non-error output')
  .option('--verbose', 'Detailed output')
  .option('--json', 'JSON output')
  .option('--no-color', 'Disable color output')
  .option('--on-error <mode>', 'Error handling: skip|fail|best-effort', 'skip')
  .action(async (source, opts) => {
    try {
      await convertAction(source, opts);
    } catch (error) {
      if (opts.json) {
        console.log(JSON.stringify({ success: false, error: error.message }));
      } else {
        console.error(chalk.red('Error:'), error.message);
      }
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Register shorthand commands dynamically ──────────────────────────
// Hidden from top-level help to keep it readable. Use `hamlet shorthands` to see them.
for (const [alias, { from, to }] of Object.entries(SHORTHANDS)) {
  program
    .command(`${alias} <source>`, { hidden: true })
    .description(`Convert ${from} \u2192 ${to}`)
    .option('-o, --output <path>', 'Output path')
    .option('-q, --quiet', 'Suppress output')
    .option('--dry-run', 'Preview without writing')
    .option('--plan', 'Show structured conversion plan')
    .option(
      '--on-error <mode>',
      'Error handling: skip|fail|best-effort',
      'skip'
    )
    .option('--json', 'JSON output')
    .option('--verbose', 'Detailed output')
    .option('--no-color', 'Disable color output')
    .action(async (source, opts) => {
      try {
        await convertAction(source, { ...opts, from, to });
      } catch (error) {
        if (opts.json) {
          console.log(JSON.stringify({ success: false, error: error.message }));
        } else {
          console.error(chalk.red('Error:'), error.message);
        }
        if (shouldShowStack()) {
          console.error(error.stack);
        }
        process.exit(1);
      }
    });
}

// ── Convert config command ───────────────────────────────────────────
program
  .command('convert-config')
  .description('Convert a test framework configuration file')
  .argument('<source>', 'Source config file path')
  .option(
    '-f, --from <framework>',
    'Source framework (auto-detected from filename if omitted)'
  )
  .option('-t, --to <framework>', 'Target framework (required)')
  .option(
    '-o, --output <path>',
    'Output file path (prints to stdout if omitted)'
  )
  .option('--dry-run', 'Preview without writing')
  .action(async (source, options) => {
    try {
      const { FileClassifier } = await import('../src/core/FileClassifier.js');
      const { ConfigConverter } = await import(
        '../src/core/ConfigConverter.js'
      );

      const toFramework = options.to;
      if (!toFramework) {
        console.error(chalk.red('Error: --to <framework> is required'));
        process.exit(2);
      }

      const content = await fs.readFile(source, 'utf8');

      // Auto-detect source framework from filename if not specified
      let fromFramework = options.from;
      if (!fromFramework) {
        const classifier = new FileClassifier();
        const classification = classifier.classify(source, content);
        if (classification.framework) {
          fromFramework = classification.framework;
          console.log(
            chalk.yellow(`Auto-detected source framework: ${fromFramework}`)
          );
        } else {
          console.error(
            chalk.red('Error:'),
            'Could not auto-detect source framework. Use --from <framework>.'
          );
          console.error(
            chalk.gray('Next steps: Specify --from <framework> explicitly.')
          );
          process.exit(2);
        }
      }
      const converter = new ConfigConverter();
      const result = converter.convert(
        content,
        fromFramework.toLowerCase(),
        toFramework.toLowerCase()
      );

      if (options.dryRun) {
        console.log(chalk.yellow('Dry run mode - no files will be modified\n'));
        console.log(`  Source: ${source}`);
        console.log(`  Detected framework: ${fromFramework}`);
        console.log(`  Target framework: ${toFramework.toLowerCase()}`);
        console.log(`  Output: ${options.output || '(stdout)'}`);
        return;
      }

      if (options.output) {
        await fs.mkdir(path.dirname(path.resolve(options.output)), {
          recursive: true,
        });
        await fs.writeFile(options.output, result);
        console.error(
          chalk.green(`Config converted: ${source} -> ${options.output}`)
        );
      } else {
        process.stdout.write(result);
      }
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── List command — categorized conversion directions ─────────────────
program
  .command('list')
  .description('List all supported conversion directions with shorthands')
  .action(() => {
    console.log(chalk.blue('\nSupported conversion directions:\n'));

    for (const category of CONVERSION_CATEGORIES) {
      console.log(chalk.bold(`  ${category.name}`));
      for (const dir of category.directions) {
        const shortcuts = dir.shorthands.join(', ');
        console.log(
          `    ${chalk.green(dir.from.padEnd(14))} ${chalk.gray('\u2192')} ${chalk.cyan(dir.to.padEnd(14))} ${chalk.gray(shortcuts)}`
        );
      }
      console.log();
    }

    console.log(
      chalk.gray(
        'Usage: hamlet convert <source> --from <framework> --to <framework>'
      )
    );
    console.log(chalk.gray('  or:  hamlet <shorthand> <source> -o <output>'));
    console.log();
  });

// ── List-conversions (backward compat, alias for list) ───────────────
program
  .command('list-conversions')
  .description('List all supported conversion directions')
  .action(() => {
    console.log(chalk.blue('\nSupported conversion directions:\n'));
    const conversions = ConverterFactory.getSupportedConversions();
    conversions.forEach((conv) => {
      const [from, to] = conv.split('-');
      console.log(
        `  ${chalk.green(from.padEnd(12))} ${chalk.gray('->')} ${chalk.cyan(to)}`
      );
    });
    console.log();
    console.log(
      chalk.gray(
        'Usage: hamlet convert <source> --from <framework> --to <framework>'
      )
    );
    console.log();
  });

// ── Shorthands command — flat alias table ────────────────────────────
program
  .command('shorthands')
  .description('List all shorthand command aliases')
  .action(() => {
    console.log(chalk.blue('\nShorthand command aliases:\n'));
    console.log(
      `  ${chalk.bold('Alias'.padEnd(18))} ${chalk.bold('From'.padEnd(14))} ${chalk.bold('To'.padEnd(14))}`
    );
    console.log(`  ${'─'.repeat(18)} ${'─'.repeat(14)} ${'─'.repeat(14)}`);

    for (const [alias, { from, to }] of Object.entries(SHORTHANDS)) {
      console.log(
        `  ${chalk.cyan(alias.padEnd(18))} ${chalk.green(from.padEnd(14))} ${chalk.green(to)}`
      );
    }
    console.log();
    console.log(chalk.gray('Usage: hamlet <shorthand> <source> -o <output>'));
    console.log();
  });

// ── Detect framework ─────────────────────────────────────────────────
program
  .command('detect')
  .description('Auto-detect the testing framework from a file')
  .argument('<file>', 'Test file to analyze')
  .action(async (file) => {
    try {
      const content = await fs.readFile(file, 'utf8');
      const result = FrameworkDetector.detect(content, file);

      console.log(chalk.blue('\nFramework Detection Results:\n'));
      console.log(`  File: ${chalk.cyan(file)}`);
      console.log(
        `  Detected Framework: ${chalk.green(result.framework || 'Unknown')}`
      );
      console.log(
        `  Confidence: ${chalk.yellow(Math.round(result.confidence * 100) + '%')}`
      );
      console.log(`  Detection Method: ${result.method}`);

      if (result.contentAnalysis?.scores) {
        console.log('\n  Scores:');
        for (const [framework, score] of Object.entries(
          result.contentAnalysis.scores
        )) {
          const bar = '\u2588'.repeat(Math.min(20, Math.round(score / 2)));
          console.log(`    ${framework.padEnd(12)} ${bar} (${score})`);
        }
      }
      console.log();
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Validate command ─────────────────────────────────────────────────
program
  .command('validate')
  .description('Validate converted tests')
  .argument('<path>', 'Path to converted tests')
  .option(
    '--framework <framework>',
    'Target framework for validation',
    'playwright'
  )
  .option('--report <format>', 'Validation report format')
  .action(async (testPath, options) => {
    try {
      console.log(chalk.blue(`Validating ${options.framework} tests...`));
      const validator = new TestValidator();
      const results = await validator.validateConvertedTests(testPath);

      if (options.report) {
        await generateReport(testPath, options.report, results);
      }

      console.log(chalk.green('Validation completed!'));
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Init command ─────────────────────────────────────────────────────
program
  .command('init')
  .description('Initialize Hamlet configuration')
  .option('-f, --force', 'Overwrite existing configuration')
  .action(async (options) => {
    try {
      console.log(chalk.blue('Initializing Hamlet configuration...'));

      const configPath = '.hamletrc.json';

      if (!options.force) {
        try {
          await fs.access(configPath);
          console.error(
            chalk.red('Error:'),
            'Configuration file already exists. Use --force to overwrite.'
          );
          console.error(
            chalk.gray(
              'Next steps: Use --force to overwrite, or edit the existing file.'
            )
          );
          process.exit(2);
        } catch {
          // File doesn't exist, continue
        }
      }

      const config = {
        defaultSource: 'cypress',
        defaultTarget: 'playwright',
        output: './converted',
        preserveStructure: true,
        validate: true,
        report: 'json',
        batchSize: 5,
        ignore: ['node_modules/**', '**/fixtures/**'],
      };

      await fs.writeFile(configPath, JSON.stringify(config, null, 2));
      console.log(chalk.green(`Configuration saved to ${configPath}`));
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Migrate command — full project migration ─────────────────────────
program
  .command('migrate')
  .description('Migrate an entire project from one test framework to another')
  .argument('<dir>', 'Project directory to migrate')
  .option(
    '-f, --from <framework>',
    'Source framework (jest, cypress, playwright)',
    'jest'
  )
  .option(
    '-t, --to <framework>',
    'Target framework (vitest, playwright, cypress)',
    'vitest'
  )
  .option('-o, --output <path>', 'Output directory for converted files')
  .option('--continue', 'Resume a previously started migration')
  .option('--retry-failed', 'Retry only previously failed files')
  .option('--dry-run', 'Preview migration without making changes')
  .option('--plan', 'Show structured migration plan')
  .action(async (dir, options) => {
    try {
      if (options.dryRun || options.plan) {
        // Migrate dry-run / plan delegates to estimator
        const { MigrationEstimator } = await import(
          '../src/core/MigrationEstimator.js'
        );
        const estimator = new MigrationEstimator();

        const result = await estimator.estimate(dir, {
          from: options.from,
          to: options.to,
        });

        if (options.plan) {
          // ── Plan output (migrate) ──────────────────────────────
          console.log(
            chalk.bold(`Migration Plan: ${options.from} \u2192 ${options.to}`)
          );
          console.log(`  Directory: ${dir}\n`);

          console.log('  Files:');
          console.log(`    Test files:   ${result.summary.testFiles}`);
          console.log(`    Config files: ${result.summary.configFiles}`);
          console.log(`    Helper files: ${result.summary.helperFiles}`);

          if (result.files && result.files.length > 0) {
            console.log(
              `\n  ${'Input'.padEnd(30)} ${'Type'.padEnd(12)} ${'Confidence'}`
            );
            for (const f of result.files) {
              const conf = f.predictedConfidence || 0;
              const level = conf >= 80 ? 'high' : conf >= 50 ? 'medium' : 'low';
              console.log(
                `  ${(f.path || '').padEnd(30)} ${(f.type || 'unknown').padEnd(12)} ${level}`
              );
            }
          }

          console.log(
            `\n  Summary: ${result.summary.predictedHigh} high, ${result.summary.predictedMedium} medium, ${result.summary.predictedLow} low`
          );
          console.log('  Warnings: (none)');
        } else {
          // ── Standard dry-run output (migrate) ──────────────────
          console.log(
            chalk.yellow('Dry run mode - no files will be modified\n')
          );
          console.log(
            chalk.blue(`Estimating migration for ${chalk.bold(dir)}...`)
          );

          console.log(chalk.bold('\nEstimation Summary:'));
          console.log(`  Total files: ${result.summary.totalFiles}`);
          console.log(`  Test files: ${result.summary.testFiles}`);
          console.log(`  Helper files: ${result.summary.helperFiles}`);
          console.log(`  Config files: ${result.summary.configFiles}`);
          console.log(
            `  ${chalk.green('High confidence:')} ${result.summary.predictedHigh}`
          );
          console.log(
            `  ${chalk.yellow('Medium confidence:')} ${result.summary.predictedMedium}`
          );
          console.log(
            `  ${chalk.red('Low confidence:')} ${result.summary.predictedLow}`
          );
        }
        return;
      }

      const { MigrationEngine } = await import(
        '../src/core/MigrationEngine.js'
      );
      const engine = new MigrationEngine();

      console.log(
        chalk.blue(
          `Migrating ${chalk.bold(dir)} from ${chalk.bold(options.from)} to ${chalk.bold(options.to)}...`
        )
      );

      const { results, checklist, state } = await engine.migrate(dir, {
        from: options.from,
        to: options.to,
        output: options.output,
        continue: options.continue,
        retryFailed: options.retryFailed,
        onProgress: (file, status, confidence) => {
          const icon =
            status === 'converted'
              ? chalk.green('\u2713')
              : status === 'skipped'
                ? chalk.yellow('\u2192')
                : status === 'failed'
                  ? chalk.red('\u2717')
                  : chalk.gray('\u00b7');
          const confStr = confidence != null ? ` (${confidence}%)` : '';
          console.log(`  ${icon} ${file}${confStr}`);
        },
      });

      console.log(
        chalk.green(
          `\nMigration complete: ${state.converted} converted, ${state.failed} failed, ${state.skipped || 0} skipped`
        )
      );

      if (state.failed > 0) {
        process.exit(state.converted > 0 ? 3 : 1);
      }
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Estimate command ─────────────────────────────────────────────────
program
  .command('estimate')
  .description('Estimate migration complexity without converting')
  .argument('<dir>', 'Project directory to estimate')
  .option('-f, --from <framework>', 'Source framework', 'jest')
  .option('-t, --to <framework>', 'Target framework', 'vitest')
  .action(async (dir, options) => {
    try {
      const { MigrationEstimator } = await import(
        '../src/core/MigrationEstimator.js'
      );
      const estimator = new MigrationEstimator();

      console.log(chalk.blue(`Estimating migration for ${chalk.bold(dir)}...`));

      const result = await estimator.estimate(dir, {
        from: options.from,
        to: options.to,
      });

      console.log(chalk.bold('\nEstimation Summary:'));
      console.log(`  Total files: ${result.summary.totalFiles}`);
      console.log(`  Test files: ${result.summary.testFiles}`);
      console.log(`  Helper files: ${result.summary.helperFiles}`);
      console.log(`  Config files: ${result.summary.configFiles}`);
      console.log(
        `  ${chalk.green('High confidence:')} ${result.summary.predictedHigh}`
      );
      console.log(
        `  ${chalk.yellow('Medium confidence:')} ${result.summary.predictedMedium}`
      );
      console.log(
        `  ${chalk.red('Low confidence:')} ${result.summary.predictedLow}`
      );

      if (result.blockers.length > 0) {
        console.log(chalk.bold('\nTop Blockers:'));
        for (const b of result.blockers) {
          console.log(
            `  ${chalk.red(b.pattern)} \u2014 ${b.count} occurrences`
          );
        }
      }

      console.log(chalk.bold('\nEffort Estimate:'));
      console.log(`  ${result.estimatedEffort.description}`);
      if (result.estimatedEffort.estimatedManualMinutes > 0) {
        console.log(
          `  Estimated manual time: ~${result.estimatedEffort.estimatedManualMinutes} minutes`
        );
      }
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Status command ───────────────────────────────────────────────────
program
  .command('status')
  .description('Show current migration progress')
  .option('-d, --dir <path>', 'Project directory', '.')
  .action(async (options) => {
    try {
      const { MigrationStateManager } = await import(
        '../src/core/MigrationStateManager.js'
      );
      const stateManager = new MigrationStateManager(path.resolve(options.dir));

      if (!(await stateManager.exists())) {
        console.log(
          chalk.yellow(
            'No migration in progress. Run `hamlet migrate` to start.'
          )
        );
        return;
      }

      const state = await stateManager.load();
      const status = stateManager.getStatus();

      console.log(chalk.bold('Migration Status:'));
      console.log(`  Source: ${chalk.cyan(status.source || 'unknown')}`);
      console.log(`  Target: ${chalk.cyan(status.target || 'unknown')}`);
      console.log(`  Started: ${status.startedAt || 'unknown'}`);
      console.log(`  Converted: ${chalk.green(status.converted)}`);
      console.log(`  Failed: ${chalk.red(status.failed)}`);
      console.log(`  Skipped: ${chalk.yellow(status.skipped)}`);
      console.log(`  Total tracked: ${status.total}`);
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Checklist command ────────────────────────────────────────────────
program
  .command('checklist')
  .description('Generate or display migration checklist')
  .option('-d, --dir <path>', 'Project directory', '.')
  .action(async (options) => {
    try {
      const { MigrationStateManager } = await import(
        '../src/core/MigrationStateManager.js'
      );
      const { MigrationChecklistGenerator } = await import(
        '../src/core/MigrationChecklistGenerator.js'
      );

      const stateManager = new MigrationStateManager(path.resolve(options.dir));

      if (!(await stateManager.exists())) {
        console.log(
          chalk.yellow(
            'No migration in progress. Run `hamlet migrate` to start.'
          )
        );
        return;
      }

      const state = await stateManager.load();
      const generator = new MigrationChecklistGenerator();

      const results = Object.entries(state.files).map(([filePath, info]) => ({
        path: filePath,
        confidence: info.confidence || 0,
        status: info.status,
        error: info.error || null,
        warnings: [],
        todos: [],
        type: 'unknown',
      }));

      const checklist = generator.generate(
        { nodes: [], edges: new Map() },
        results
      );
      console.log(checklist);
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Reset command ────────────────────────────────────────────────────
program
  .command('reset')
  .description('Clear migration state (.hamlet/ directory)')
  .option('-d, --dir <path>', 'Project directory', '.')
  .option('-y, --yes', 'Skip confirmation prompt')
  .action(async (options) => {
    try {
      const { MigrationStateManager } = await import(
        '../src/core/MigrationStateManager.js'
      );
      const stateManager = new MigrationStateManager(path.resolve(options.dir));

      if (!(await stateManager.exists())) {
        console.log(chalk.yellow('No migration state to reset.'));
        return;
      }

      if (!options.yes) {
        console.log(
          chalk.yellow(
            'This will remove the .hamlet/ directory and all migration state.'
          )
        );
        console.log(chalk.yellow('Use --yes to confirm.'));
        return;
      }

      await stateManager.reset();
      console.log(chalk.green('Migration state cleared.'));
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

// ── Doctor command — diagnostics ─────────────────────────────────────
program
  .command('doctor')
  .description('Run diagnostics and check Hamlet setup')
  .argument('[path]', 'Directory to diagnose', '.')
  .option('--json', 'JSON output')
  .option('--verbose', 'Show additional detail for each check')
  .action(async (targetPath, options) => {
    try {
      const { runDoctor } = await import('../src/cli/doctor.js');
      const result = await runDoctor(targetPath);
      const isVerbose = options.verbose || program.opts().verbose;

      if (options.json) {
        const output = {
          checks: result.checks.map((c) => {
            const obj = {
              id: c.id,
              label: c.label,
              status: c.status,
              detail: c.detail,
            };
            if (c.remediation) obj.remediation = c.remediation;
            if (isVerbose && c.verbose) obj.verbose = c.verbose;
            return obj;
          }),
          summary: result.summary,
        };
        console.log(JSON.stringify(output, null, 2));
      } else {
        console.log(chalk.blue('\nHamlet Doctor\n'));
        for (const check of result.checks) {
          const tag =
            check.status === 'PASS'
              ? chalk.green('[PASS]')
              : check.status === 'WARN'
                ? chalk.yellow('[WARN]')
                : chalk.red('[FAIL]');
          console.log(`  ${tag} ${check.label}: ${check.detail}`);
          if (isVerbose && check.verbose) {
            console.log(`         ${chalk.dim(check.verbose)}`);
          }
          if (check.remediation) {
            console.log(`         ${chalk.yellow('→')} ${check.remediation}`);
          }
        }
        const { pass, warn, fail, total } = result.summary;
        console.log(
          `\n  ${total} checks: ${chalk.green(pass + ' passed')}` +
            (warn ? `, ${chalk.yellow(warn + ' warnings')}` : '') +
            (fail ? `, ${chalk.red(fail + ' failed')}` : '')
        );
        console.log();
      }

      if (result.hasFail) process.exit(1);
    } catch (error) {
      console.error(chalk.red('Error:'), error.message);
      if (shouldShowStack()) console.error(error.stack);
      process.exit(1);
    }
  });

program.parse();
