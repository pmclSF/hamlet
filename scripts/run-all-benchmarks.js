#!/usr/bin/env node
/**
 * Hamlet Benchmark Matrix Runner
 *
 * Runs all benchmark conversion pairs and produces a coverage summary.
 *
 * Usage:
 *   node scripts/run-all-benchmarks.js              # run all pairs (--max 50)
 *   node scripts/run-all-benchmarks.js --category python
 *   node scripts/run-all-benchmarks.js --max 30
 *   node scripts/run-all-benchmarks.js --summary     # aggregate existing results
 */

import { execFile } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const execFileAsync = promisify(execFile);
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '..');
const benchmarkScript = path.join(rootDir, 'benchmarks', 'benchmark.js');
const resultsDir = path.join(rootDir, 'benchmarks', 'results');

// ── Benchmark matrix ────────────────────────────────────────────────
const BENCHMARK_MATRIX = [
  // JS Unit Testing
  {
    category: 'js-unit',
    from: 'jest',
    to: 'mocha',
    dirs: ['benchmarks/jest-vitest/'],
    label: 'jest->mocha (Vue, Prettier, etc.)',
  },
  {
    category: 'js-unit',
    from: 'jest',
    to: 'jasmine',
    dirs: ['benchmarks/jest-vitest/'],
    label: 'jest->jasmine',
  },
  {
    category: 'js-unit',
    from: 'jest',
    to: 'vitest',
    dirs: ['benchmarks/jest-vitest/'],
    label: 'jest->vitest',
    forwardOnly: true, // no vitest->jest converter for round-trip
  },
  {
    category: 'js-unit',
    from: 'mocha',
    to: 'jest',
    dirs: [
      'benchmarks/mocha-jasmine-etc/express/test',
      'benchmarks/mocha-jasmine-etc/chai',
      'benchmarks/mocha-jasmine-etc/mongoose/test',
    ],
    label: 'mocha->jest (Express, Chai, Mongoose)',
  },
  {
    category: 'js-unit',
    from: 'jasmine',
    to: 'jest',
    dirs: ['benchmarks/mocha-jasmine-etc/jasmine/spec'],
    label: 'jasmine->jest (Jasmine core)',
  },
  // Note: vitest repos (pinia, vite, unocss) are downloaded for reference
  // validation but cannot be used as benchmark inputs because no vitest->*
  // converter exists. The jest->vitest direction is tested above.

  // E2E
  {
    category: 'e2e',
    from: 'cypress',
    to: 'playwright',
    dirs: ['benchmarks/cypress-playwright/'],
    label: 'cypress->playwright (mixed repos)',
  },
  {
    category: 'e2e',
    from: 'playwright',
    to: 'cypress',
    dirs: ['benchmarks/cypress-playwright/nocodb/'],
    label: 'playwright->cypress (NocoDB)',
  },
  {
    category: 'e2e',
    from: 'cypress',
    to: 'webdriverio',
    dirs: ['benchmarks/cypress-playwright/cypress-example-kitchensink/'],
    label: 'cypress->webdriverio (Kitchensink)',
  },
  {
    category: 'e2e',
    from: 'cypress',
    to: 'selenium',
    dirs: ['benchmarks/cypress-playwright/cypress-example-kitchensink/'],
    label: 'cypress->selenium (Kitchensink)',
  },
  {
    category: 'e2e',
    from: 'puppeteer',
    to: 'playwright',
    dirs: ['benchmarks/mocha-jasmine-etc/puppeteer/test/src'],
    label: 'puppeteer->playwright',
  },
  {
    category: 'e2e',
    from: 'playwright',
    to: 'puppeteer',
    dirs: ['benchmarks/cypress-playwright/nocodb/'],
    label: 'playwright->puppeteer (NocoDB)',
  },
  {
    category: 'e2e',
    from: 'webdriverio',
    to: 'playwright',
    dirs: ['benchmarks/mocha-jasmine-etc/webdriverio/'],
    label: 'webdriverio->playwright',
  },
  {
    category: 'e2e',
    from: 'testcafe',
    to: 'playwright',
    dirs: ['benchmarks/mocha-jasmine-etc/testcafe/'],
    label: 'testcafe->playwright',
    forwardOnly: true, // no playwright->testcafe converter
  },
  {
    category: 'e2e',
    from: 'testcafe',
    to: 'cypress',
    dirs: ['benchmarks/mocha-jasmine-etc/testcafe/'],
    label: 'testcafe->cypress',
    forwardOnly: true, // no cypress->testcafe converter
  },

  // Java
  {
    category: 'java',
    from: 'junit5',
    to: 'testng',
    dirs: ['benchmarks/java/commons-lang/src/test/'],
    label: 'junit5->testng (commons-lang)',
  },
  {
    category: 'java',
    from: 'testng',
    to: 'junit5',
    dirs: ['benchmarks/java/testng/'],
    label: 'testng->junit5',
  },
  {
    category: 'java',
    from: 'junit4',
    to: 'junit5',
    dirs: [
      'benchmarks/java/mockito/src/test',
      'benchmarks/java/guava',
    ],
    label: 'junit4->junit5 (Mockito, Guava)',
    forwardOnly: true, // no junit5->junit4 converter exists
  },

  // Python
  {
    category: 'python',
    from: 'pytest',
    to: 'unittest',
    dirs: [
      'benchmarks/python/pytest/testing',
      'benchmarks/python/flask/tests',
      'benchmarks/python/requests/tests',
      'benchmarks/python/httpx/tests',
    ],
    label: 'pytest->unittest (pytest, Flask, Requests, httpx)',
  },
  {
    category: 'python',
    from: 'unittest',
    to: 'pytest',
    dirs: ['benchmarks/python/django/tests/'],
    label: 'unittest->pytest (Django)',
  },
];

// ── Argument parsing ────────────────────────────────────────────────
const args = process.argv.slice(2);
let maxFiles = 50;
let categoryFilter = null;
let summaryOnly = false;

for (let i = 0; i < args.length; i++) {
  if (args[i] === '--max' && args[i + 1]) maxFiles = parseInt(args[++i]);
  if (args[i] === '--category' && args[i + 1]) categoryFilter = args[++i];
  if (args[i] === '--summary') summaryOnly = true;
}

// ── Summary mode: aggregate existing result files ───────────────────
async function printSummary() {
  let files;
  try {
    files = await fs.readdir(resultsDir);
  } catch {
    console.error('No results directory found. Run benchmarks first.');
    process.exit(1);
  }

  // Group by direction, keeping only the most recent per direction
  const latest = new Map();
  for (const f of files) {
    if (!f.endsWith('.json')) continue;
    // Format: from-to-roundtrip-TIMESTAMP.json
    const match = f.match(/^(.+)-roundtrip-(.+)\.json$/);
    if (!match) continue;
    const direction = match[1];
    const timestamp = match[2];
    if (!latest.has(direction) || timestamp > latest.get(direction).timestamp) {
      latest.set(direction, { file: f, timestamp });
    }
  }

  console.log('');
  console.log(
    '=== Hamlet Benchmark Coverage Matrix (latest results per direction) ===',
  );
  console.log('');
  console.log(
    padRight('Direction', 30) +
      padRight('Files', 8) +
      padRight('Fwd OK', 8) +
      padRight('RT OK', 8) +
      padRight('Avg Fid', 10) +
      padRight('Med Fid', 10) +
      padRight('Struct', 10) +
      'TODOs',
  );
  console.log('-'.repeat(94));

  const sortedDirs = [...latest.keys()].sort();
  const totals = {
    files: 0,
    fwdOk: 0,
    rtOk: 0,
    fidelitySum: 0,
    structSum: 0,
    count: 0,
    todos: 0,
  };

  for (const dir of sortedDirs) {
    const { file } = latest.get(dir);
    try {
      const data = JSON.parse(
        await fs.readFile(path.join(resultsDir, file), 'utf8'),
      );
      const avgFid =
        typeof data.avgFidelity === 'number'
          ? data.avgFidelity.toFixed(4)
          : data.avgFidelity || 'N/A';
      const medFid =
        typeof data.medianFidelity === 'number'
          ? data.medianFidelity.toFixed(4)
          : 'N/A';
      const struct =
        typeof data.avgStructuralFidelity === 'number'
          ? data.avgStructuralFidelity.toFixed(4)
          : 'N/A';

      console.log(
        padRight(dir, 30) +
          padRight(String(data.totalFiles || 0), 8) +
          padRight(String(data.forwardOk || 0), 8) +
          padRight(String(data.roundtripOk || 0), 8) +
          padRight(avgFid, 10) +
          padRight(medFid, 10) +
          padRight(struct, 10) +
          (data.todoCount || 0),
      );

      totals.files += data.totalFiles || 0;
      totals.fwdOk += data.forwardOk || 0;
      totals.rtOk += data.roundtripOk || 0;
      if (typeof data.avgFidelity === 'number') {
        totals.fidelitySum += data.avgFidelity;
        totals.count++;
      }
      if (typeof data.avgStructuralFidelity === 'number') {
        totals.structSum += data.avgStructuralFidelity;
      }
      totals.todos += data.todoCount || 0;
    } catch {
      console.log(padRight(dir, 30) + '  (failed to read result file)');
    }
  }

  console.log('-'.repeat(94));
  console.log(
    padRight('TOTAL', 30) +
      padRight(String(totals.files), 8) +
      padRight(String(totals.fwdOk), 8) +
      padRight(String(totals.rtOk), 8) +
      padRight(
        totals.count > 0 ? (totals.fidelitySum / totals.count).toFixed(4) : 'N/A',
        10,
      ) +
      padRight('', 10) +
      padRight(
        totals.count > 0 ? (totals.structSum / totals.count).toFixed(4) : 'N/A',
        10,
      ) +
      totals.todos,
  );
  console.log('');
  console.log(`${sortedDirs.length} directions with results`);
}

function padRight(str, len) {
  return String(str).padEnd(len);
}

// ── Run a single benchmark entry ────────────────────────────────────
async function runBenchmarkEntry(entry, max) {
  const results = [];

  for (const dir of entry.dirs) {
    const absDir = path.resolve(rootDir, dir);

    // Check if directory exists
    try {
      await fs.access(absDir);
    } catch {
      console.log(`  SKIP  ${dir} (not found — run download-benchmarks.sh)`);
      results.push({ dir, status: 'skipped', reason: 'dir-not-found' });
      continue;
    }

    const mode = entry.forwardOnly ? 'forward-only' : 'round-trip';
    console.log(`  RUN   ${entry.from}->${entry.to} on ${dir} (max ${max}, ${mode})`);

    try {
      const benchArgs = [benchmarkScript, absDir, entry.from, entry.to, '--max', String(max)];
      if (entry.forwardOnly) benchArgs.push('--forward-only');

      const { stdout, stderr } = await execFileAsync(
        'node',
        benchArgs,
        {
          cwd: rootDir,
          timeout: 300_000, // 5 min per entry
          maxBuffer: 10 * 1024 * 1024,
        },
      );

      // Parse summary from stdout
      const avgMatch = stdout.match(/Avg line fidelity:\s+([\d.]+)/);
      const medMatch = stdout.match(/Median line fidelity:\s+([\d.]+)/);
      const totalMatch = stdout.match(/Total files:\s+(\d+)/);
      const fwdMatch = stdout.match(/Forward OK:\s+(\d+)/);
      const rtMatch = stdout.match(/Round-trip OK:\s+(\d+)/);

      results.push({
        dir,
        status: 'ok',
        totalFiles: totalMatch ? parseInt(totalMatch[1]) : 0,
        forwardOk: fwdMatch ? parseInt(fwdMatch[1]) : 0,
        roundtripOk: rtMatch ? parseInt(rtMatch[1]) : 0,
        avgFidelity: avgMatch ? parseFloat(avgMatch[1]) : 0,
        medianFidelity: medMatch ? parseFloat(medMatch[1]) : 0,
      });

      // Print last few lines of output (the summary)
      const lines = stdout.trim().split('\n');
      const summaryStart = lines.findIndex((l) => l.includes('=== Summary ==='));
      if (summaryStart >= 0) {
        for (const line of lines.slice(summaryStart)) {
          console.log(`        ${line}`);
        }
      }
    } catch (e) {
      const msg = e.stderr
        ? e.stderr.trim().split('\n').slice(-2).join(' ')
        : e.message;
      console.log(`  FAIL  ${dir}: ${msg}`);
      results.push({ dir, status: 'error', error: msg });
    }
  }

  return results;
}

// ── Main ────────────────────────────────────────────────────────────
async function main() {
  if (summaryOnly) {
    await printSummary();
    return;
  }

  const entries = categoryFilter
    ? BENCHMARK_MATRIX.filter((e) => e.category === categoryFilter)
    : BENCHMARK_MATRIX;

  if (entries.length === 0) {
    const categories = [...new Set(BENCHMARK_MATRIX.map((e) => e.category))];
    console.error(
      `Unknown category: ${categoryFilter}. Available: ${categories.join(', ')}`,
    );
    process.exit(1);
  }

  console.log('=== Hamlet Full Benchmark Run ===');
  console.log(`Entries: ${entries.length}`);
  console.log(`Max files per dir: ${maxFiles}`);
  console.log('');

  const allResults = [];

  for (const entry of entries) {
    console.log(`\n--- ${entry.label} ---`);
    const results = await runBenchmarkEntry(entry, maxFiles);
    allResults.push({ ...entry, results });
  }

  // Print summary table
  console.log('\n');
  console.log('=== Run Summary ===');
  console.log('');
  console.log(
    padRight('Direction', 35) +
      padRight('Files', 8) +
      padRight('Fwd OK', 8) +
      padRight('RT OK', 8) +
      padRight('Avg Fid', 10) +
      'Status',
  );
  console.log('-'.repeat(79));

  for (const entry of allResults) {
    const okResults = entry.results.filter((r) => r.status === 'ok');
    const skipped = entry.results.filter((r) => r.status === 'skipped');
    const errors = entry.results.filter((r) => r.status === 'error');

    if (okResults.length === 0) {
      const reason = skipped.length > 0 ? 'skipped' : 'error';
      console.log(
        padRight(`${entry.from}->${entry.to}`, 35) +
          padRight('-', 8) +
          padRight('-', 8) +
          padRight('-', 8) +
          padRight('-', 10) +
          reason,
      );
      continue;
    }

    const totalFiles = okResults.reduce((s, r) => s + r.totalFiles, 0);
    const fwdOk = okResults.reduce((s, r) => s + r.forwardOk, 0);
    const rtOk = okResults.reduce((s, r) => s + r.roundtripOk, 0);
    const avgFid =
      okResults.reduce((s, r) => s + r.avgFidelity, 0) / okResults.length;

    let status = 'ok';
    if (errors.length > 0) status += ` (${errors.length} errors)`;
    if (skipped.length > 0) status += ` (${skipped.length} skipped)`;

    console.log(
      padRight(`${entry.from}->${entry.to}`, 35) +
        padRight(String(totalFiles), 8) +
        padRight(String(fwdOk), 8) +
        padRight(String(rtOk), 8) +
        padRight(avgFid.toFixed(4), 10) +
        status,
    );
  }

  console.log('');
  console.log('Done. Run with --summary to see aggregated latest results.');
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
