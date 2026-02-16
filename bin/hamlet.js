#!/usr/bin/env node

import { program } from 'commander';
import chalk from 'chalk';
import { fileURLToPath } from 'url';
import { dirname } from 'path';
import { createRequire } from 'module';
import fs from 'fs/promises';
import path from 'path';
import {
  convertFile,
  convertRepository,
  validateTests,
  generateReport,
  TestValidator,
  ConversionReporter
} from '../src/index.js';
import { ConverterFactory, FRAMEWORKS } from '../src/core/ConverterFactory.js';
import { FrameworkDetector } from '../src/core/FrameworkDetector.js';

const require = createRequire(import.meta.url);
const version = require('../package.json').version;

program
  .version(version)
  .description('Hamlet: Bidirectional multi-framework test converter for Cypress, Playwright, Selenium, Jest, and Vitest.');

// Main convert command with --from and --to flags
program
  .command('convert')
  .description('Convert tests between frameworks (Cypress, Playwright, Selenium, Jest, Vitest)')
  .argument('<source>', 'Source test file, directory, or repository URL')
  .option('-f, --from <framework>', 'Source framework (cypress, playwright, selenium, jest, vitest)', 'cypress')
  .option('-t, --to <framework>', 'Target framework (cypress, playwright, selenium, jest, vitest)', 'playwright')
  .option('-o, --output <path>', 'Output path for converted tests')
  .option('-c, --config <path>', 'Custom configuration file path')
  .option('--test-type <type>', 'Specify test type (e2e, component, api, etc.)')
  .option('--validate', 'Validate converted tests')
  .option('--report <format>', 'Generate conversion report (html, json, markdown)')
  .option('--preserve-structure', 'Maintain original directory structure')
  .option('--batch-size <number>', 'Number of tests to process in parallel', '5')
  .option('--ignore <pattern>', 'Files to ignore (glob pattern)')
  .option('--dry-run', 'Show what would be converted without making changes')
  .option('--auto-detect', 'Auto-detect source framework from file content')
  .action(async (source, options) => {
    try {
      // Validate frameworks
      const validFrameworks = Object.values(FRAMEWORKS);
      const fromFramework = options.from.toLowerCase();
      const toFramework = options.to.toLowerCase();

      if (!validFrameworks.includes(fromFramework)) {
        console.error(chalk.red(`Invalid source framework: ${options.from}`));
        console.error(`Valid options: ${validFrameworks.join(', ')}`);
        process.exit(1);
      }

      if (!validFrameworks.includes(toFramework)) {
        console.error(chalk.red(`Invalid target framework: ${options.to}`));
        console.error(`Valid options: ${validFrameworks.join(', ')}`);
        process.exit(1);
      }

      if (fromFramework === toFramework) {
        console.error(chalk.red('Source and target frameworks must be different'));
        process.exit(1);
      }

      console.log(chalk.blue(`Converting from ${chalk.bold(fromFramework)} to ${chalk.bold(toFramework)}...`));

      // Auto-detect framework if requested
      if (options.autoDetect) {
        try {
          const content = await fs.readFile(source, 'utf8');
          const detection = FrameworkDetector.detectFromContent(content);
          if (detection.framework && detection.confidence > 0.5) {
            console.log(chalk.yellow(`Auto-detected source framework: ${detection.framework} (${Math.round(detection.confidence * 100)}% confidence)`));
            options.from = detection.framework;
          }
        } catch (e) {
          // Ignore auto-detection errors
        }
      }

      // Create converter
      const converter = await ConverterFactory.createConverter(fromFramework, toFramework, {
        batchSize: parseInt(options.batchSize),
        preserveStructure: options.preserveStructure
      });

      if (options.dryRun) {
        console.log(chalk.yellow('Dry run mode - no files will be modified'));
        console.log(`Would convert: ${source}`);
        console.log(`Output: ${options.output || 'same directory with new extension'}`);
        return;
      }

      const isRepository = source.includes('github.com') || source.includes('gitlab.com');

      if (isRepository) {
        await convertRepository(source, options.output, {
          ...options,
          converter,
          fromFramework,
          toFramework
        });
      } else {
        // Check if source is a file or directory
        const sourceStat = await fs.stat(source);

        if (sourceStat.isDirectory()) {
          // Convert all test files in the directory
          const files = await fs.readdir(source);
          const testFiles = files.filter(f =>
            f.endsWith('.js') || f.endsWith('.ts') || f.endsWith('.tsx') || f.endsWith('.jsx')
          );

          for (const file of testFiles) {
            const filePath = path.join(source, file);
            const content = await fs.readFile(filePath, 'utf8');
            const converted = await converter.convert(content, options);

            // Determine output filename
            const ext = path.extname(file);
            const base = path.basename(file, ext);
            let newExt;
            if (toFramework === 'cypress') {
              newExt = '.cy.js';
            } else if (toFramework === 'playwright') {
              newExt = '.spec.js';
            } else if (toFramework === 'vitest') {
              newExt = '.test.js';
            } else {
              newExt = '.test.js';
            }
            const newFilename = base.replace(/\.(cy|spec|test)$/, '') + newExt;

            const outputDir = options.output || source;
            await fs.mkdir(outputDir, { recursive: true });
            const outputPath = path.join(outputDir, newFilename);
            await fs.writeFile(outputPath, converted);

            console.log(chalk.green(`Converted: ${filePath} -> ${outputPath}`));
          }
        } else {
          // Single file conversion
          // Determine output path
          let outputPath = options.output;

          // Check if output is a directory
          if (outputPath) {
            try {
              const outputStat = await fs.stat(outputPath);
              if (outputStat.isDirectory()) {
                // Output is a directory, construct filename
                const ext = path.extname(source);
                const base = path.basename(source, ext);
                let newExt;
                if (toFramework === 'cypress') {
                  newExt = '.cy.js';
                } else if (toFramework === 'playwright') {
                  newExt = '.spec.js';
                } else {
                  newExt = '.test.js';
                }
                const newFilename = base.replace(/\.(cy|spec|test)$/, '') + newExt;
                outputPath = path.join(outputPath, newFilename);
              }
            } catch (e) {
              // Output path doesn't exist yet, check if it looks like a directory
              if (!path.extname(outputPath)) {
                // No extension, treat as directory
                const ext = path.extname(source);
                const base = path.basename(source, ext);
                let newExt;
                if (toFramework === 'cypress') {
                  newExt = '.cy.js';
                } else if (toFramework === 'playwright') {
                  newExt = '.spec.js';
                } else {
                  newExt = '.test.js';
                }
                const newFilename = base.replace(/\.(cy|spec|test)$/, '') + newExt;
                await fs.mkdir(outputPath, { recursive: true });
                outputPath = path.join(outputPath, newFilename);
              }
            }
          }

          if (!outputPath) {
            const ext = path.extname(source);
            const base = path.basename(source, ext);
            const dir = path.dirname(source);

            // Determine new extension based on target framework
            let newExt;
            if (toFramework === 'cypress') {
              newExt = '.cy' + ext;
            } else if (toFramework === 'playwright') {
              newExt = '.spec' + ext;
            } else if (toFramework === 'vitest') {
              newExt = '.test' + ext;
            } else {
              newExt = '.test' + ext;
            }

            outputPath = path.join(dir, base.replace(/\.(cy|spec|test)$/, '') + newExt);
          }

          // Read source file
          const content = await fs.readFile(source, 'utf8');

          // Convert using the converter
          const converted = await converter.convert(content, options);

          // Write output
          await fs.mkdir(path.dirname(outputPath), { recursive: true });
          await fs.writeFile(outputPath, converted);

          console.log(chalk.green(`Converted: ${source} -> ${outputPath}`));
        }
      }

      if (options.validate && options.output) {
        console.log(chalk.blue('\nValidating converted tests...'));
        await validateTests(options.output);
      }

      if (options.report && options.output) {
        console.log(chalk.blue('\nGenerating report...'));
        await generateReport(options.output, options.report);
      }

      console.log(chalk.green('\nConversion completed successfully!'));
    } catch (error) {
      console.error(chalk.red('Error during conversion:'), error.message);
      if (process.env.DEBUG) {
        console.error(error.stack);
      }
      process.exit(1);
    }
  });

// Shorthand command for Cypress to Playwright (backward compatibility)
program
  .command('cy2pw')
  .description('Convert Cypress tests to Playwright (shorthand for convert --from cypress --to playwright)')
  .argument('<source>', 'Source Cypress test file or directory')
  .option('-o, --output <path>', 'Output path for converted tests')
  .option('--validate', 'Validate converted tests')
  .action(async (source, options) => {
    // Delegate to main convert command
    await program.parseAsync([
      'node', 'hamlet', 'convert', source,
      '--from', 'cypress',
      '--to', 'playwright',
      ...(options.output ? ['-o', options.output] : []),
      ...(options.validate ? ['--validate'] : [])
    ]);
  });

// Shorthand command for Jest to Vitest
program
  .command('jest2vt')
  .description('Convert Jest tests to Vitest (shorthand for convert --from jest --to vitest)')
  .argument('<source>', 'Source Jest test file or directory')
  .option('-o, --output <path>', 'Output path for converted tests')
  .action(async (source, options) => {
    // Delegate to main convert command
    await program.parseAsync([
      'node', 'hamlet', 'convert', source,
      '--from', 'jest',
      '--to', 'vitest',
      ...(options.output ? ['-o', options.output] : [])
    ]);
  });

// List supported conversions
program
  .command('list-conversions')
  .description('List all supported conversion directions')
  .action(() => {
    console.log(chalk.blue('\nSupported conversion directions:\n'));
    const conversions = ConverterFactory.getSupportedConversions();
    conversions.forEach(conv => {
      const [from, to] = conv.split('-');
      console.log(`  ${chalk.green(from.padEnd(12))} ${chalk.gray('->')} ${chalk.cyan(to)}`);
    });
    console.log();
    console.log(chalk.gray('Usage: hamlet convert <source> --from <framework> --to <framework>'));
    console.log();
  });

// Detect framework
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
      console.log(`  Detected Framework: ${chalk.green(result.framework || 'Unknown')}`);
      console.log(`  Confidence: ${chalk.yellow(Math.round(result.confidence * 100) + '%')}`);
      console.log(`  Detection Method: ${result.method}`);

      if (result.contentAnalysis?.scores) {
        console.log('\n  Scores:');
        for (const [framework, score] of Object.entries(result.contentAnalysis.scores)) {
          const bar = '█'.repeat(Math.min(20, Math.round(score / 2)));
          console.log(`    ${framework.padEnd(12)} ${bar} (${score})`);
        }
      }
      console.log();
    } catch (error) {
      console.error(chalk.red('Error detecting framework:'), error.message);
      process.exit(1);
    }
  });

// Validate command
program
  .command('validate')
  .description('Validate converted tests')
  .argument('<path>', 'Path to converted tests')
  .option('--framework <framework>', 'Target framework for validation', 'playwright')
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
      console.error(chalk.red('Validation error:'), error.message);
      process.exit(1);
    }
  });

// Init command
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
          console.error(chalk.yellow('Configuration file already exists. Use --force to overwrite.'));
          process.exit(1);
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
        ignore: ['node_modules/**', '**/fixtures/**']
      };

      await fs.writeFile(configPath, JSON.stringify(config, null, 2));
      console.log(chalk.green(`Configuration saved to ${configPath}`));
    } catch (error) {
      console.error(chalk.red('Initialization error:'), error.message);
      process.exit(1);
    }
  });

// Migrate command — full project migration
program
  .command('migrate')
  .description('Migrate an entire project from one test framework to another')
  .argument('<dir>', 'Project directory to migrate')
  .option('-f, --from <framework>', 'Source framework (jest, cypress, playwright)', 'jest')
  .option('-t, --to <framework>', 'Target framework (vitest, playwright, cypress)', 'vitest')
  .option('-o, --output <path>', 'Output directory for converted files')
  .option('--continue', 'Resume a previously started migration')
  .option('--retry-failed', 'Retry only previously failed files')
  .action(async (dir, options) => {
    try {
      const { MigrationEngine } = await import('../src/core/MigrationEngine.js');
      const engine = new MigrationEngine();

      console.log(chalk.blue(`Migrating ${chalk.bold(dir)} from ${chalk.bold(options.from)} to ${chalk.bold(options.to)}...`));

      const { results, checklist, state } = await engine.migrate(dir, {
        from: options.from,
        to: options.to,
        output: options.output,
        continue: options.continue,
        retryFailed: options.retryFailed,
        onProgress: (file, status, confidence) => {
          const icon = status === 'converted' ? chalk.green('✓') :
                       status === 'skipped' ? chalk.yellow('→') :
                       status === 'failed' ? chalk.red('✗') : chalk.gray('·');
          const confStr = confidence != null ? ` (${confidence}%)` : '';
          console.log(`  ${icon} ${file}${confStr}`);
        },
      });

      console.log(chalk.green(`\nMigration complete: ${state.converted} converted, ${state.failed} failed, ${state.skipped || 0} skipped`));
    } catch (error) {
      console.error(chalk.red('Migration error:'), error.message);
      if (process.env.DEBUG) console.error(error.stack);
      process.exit(1);
    }
  });

// Estimate command — dry-run complexity estimate
program
  .command('estimate')
  .description('Estimate migration complexity without converting')
  .argument('<dir>', 'Project directory to estimate')
  .option('-f, --from <framework>', 'Source framework', 'jest')
  .option('-t, --to <framework>', 'Target framework', 'vitest')
  .action(async (dir, options) => {
    try {
      const { MigrationEstimator } = await import('../src/core/MigrationEstimator.js');
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
      console.log(`  ${chalk.green('High confidence:')} ${result.summary.predictedHigh}`);
      console.log(`  ${chalk.yellow('Medium confidence:')} ${result.summary.predictedMedium}`);
      console.log(`  ${chalk.red('Low confidence:')} ${result.summary.predictedLow}`);

      if (result.blockers.length > 0) {
        console.log(chalk.bold('\nTop Blockers:'));
        for (const b of result.blockers) {
          console.log(`  ${chalk.red(b.pattern)} — ${b.count} occurrences`);
        }
      }

      console.log(chalk.bold('\nEffort Estimate:'));
      console.log(`  ${result.estimatedEffort.description}`);
      if (result.estimatedEffort.estimatedManualMinutes > 0) {
        console.log(`  Estimated manual time: ~${result.estimatedEffort.estimatedManualMinutes} minutes`);
      }
    } catch (error) {
      console.error(chalk.red('Estimation error:'), error.message);
      if (process.env.DEBUG) console.error(error.stack);
      process.exit(1);
    }
  });

// Status command — show migration progress
program
  .command('status')
  .description('Show current migration progress')
  .option('-d, --dir <path>', 'Project directory', '.')
  .action(async (options) => {
    try {
      const { MigrationStateManager } = await import('../src/core/MigrationStateManager.js');
      const stateManager = new MigrationStateManager(path.resolve(options.dir));

      if (!await stateManager.exists()) {
        console.log(chalk.yellow('No migration in progress. Run `hamlet migrate` to start.'));
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
      console.error(chalk.red('Status error:'), error.message);
      process.exit(1);
    }
  });

// Checklist command — generate migration checklist
program
  .command('checklist')
  .description('Generate or display migration checklist')
  .option('-d, --dir <path>', 'Project directory', '.')
  .action(async (options) => {
    try {
      const { MigrationStateManager } = await import('../src/core/MigrationStateManager.js');
      const { MigrationChecklistGenerator } = await import('../src/core/MigrationChecklistGenerator.js');

      const stateManager = new MigrationStateManager(path.resolve(options.dir));

      if (!await stateManager.exists()) {
        console.log(chalk.yellow('No migration in progress. Run `hamlet migrate` to start.'));
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

      const checklist = generator.generate({ nodes: [], edges: new Map() }, results);
      console.log(checklist);
    } catch (error) {
      console.error(chalk.red('Checklist error:'), error.message);
      process.exit(1);
    }
  });

// Reset command — clear migration state
program
  .command('reset')
  .description('Clear migration state (.hamlet/ directory)')
  .option('-d, --dir <path>', 'Project directory', '.')
  .option('-y, --yes', 'Skip confirmation prompt')
  .action(async (options) => {
    try {
      const { MigrationStateManager } = await import('../src/core/MigrationStateManager.js');
      const stateManager = new MigrationStateManager(path.resolve(options.dir));

      if (!await stateManager.exists()) {
        console.log(chalk.yellow('No migration state to reset.'));
        return;
      }

      if (!options.yes) {
        console.log(chalk.yellow('This will remove the .hamlet/ directory and all migration state.'));
        console.log(chalk.yellow('Use --yes to confirm.'));
        return;
      }

      await stateManager.reset();
      console.log(chalk.green('Migration state cleared.'));
    } catch (error) {
      console.error(chalk.red('Reset error:'), error.message);
      process.exit(1);
    }
  });

program.parse();
