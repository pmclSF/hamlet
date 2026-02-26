#!/usr/bin/env node
/**
 * Hamlet Round-Trip Benchmark Runner
 *
 * Usage:
 *   node benchmark.js <input-dir> <from> <to> [--pattern "*.test.js"] [--max 50]
 *
 * Converts test files from->to, then to->from (round-trip).
 * Reports fidelity metrics per file and aggregate.
 */

import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import { glob } from 'fs';
import { promisify } from 'util';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '..');

// Dynamic import of Hamlet's converter
const { ConverterFactory } = await import(
  path.join(rootDir, 'src/core/ConverterFactory.js')
);

// Parse args
const args = process.argv.slice(2);
const inputDir = args[0];
const from = args[1];
const to = args[2];

let pattern = '**/*.test.js';
let maxFiles = Infinity;
let verbose = false;
let forwardOnly = false;

for (let i = 3; i < args.length; i++) {
  if (args[i] === '--pattern' && args[i + 1]) pattern = args[++i];
  if (args[i] === '--max' && args[i + 1]) maxFiles = parseInt(args[++i]);
  if (args[i] === '--verbose' || args[i] === '-v') verbose = true;
  if (args[i] === '--forward-only') forwardOnly = true;
}

if (!inputDir || !from || !to) {
  console.error(
    'Usage: node benchmark.js <input-dir> <from> <to> [--pattern "*.test.js"] [--max 50] [-v] [--forward-only]',
  );
  process.exit(1);
}

// Find test files using Node's fs.glob (Node 22+) or manual walk
async function findFiles(dir, globPattern) {
  const results = [];

  async function walk(d) {
    const entries = await fs.readdir(d, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(d, entry.name);
      if (entry.isDirectory()) {
        if (
          ['node_modules', '.git', 'dist', 'build', '__snapshots__'].includes(
            entry.name,
          )
        )
          continue;
        await walk(fullPath);
      } else if (entry.isFile()) {
        // Skip macOS resource fork files
        if (entry.name.startsWith('._')) continue;
        // Match against pattern extensions
        const ext = getExpectedExtensions(from);
        if (ext.some((e) => entry.name.endsWith(e))) {
          results.push(fullPath);
        }
      }
    }
  }

  await walk(dir);
  return results.sort();
}

function getExpectedExtensions(framework) {
  const map = {
    jest: ['.test.js', '.test.ts', '.test.jsx', '.test.tsx', '.spec.js', '.spec.ts'],
    vitest: ['.test.js', '.test.ts', '.test.jsx', '.test.tsx', '.spec.js', '.spec.ts'],
    mocha: ['.test.js', '.spec.js', '.test.ts', '.spec.ts', '.js'],
    jasmine: ['.spec.js', '.spec.ts', '_spec.js', 'Spec.js'],
    cypress: ['.cy.js', '.cy.ts', '.spec.js', '.spec.ts', '.input.js'],
    playwright: ['.spec.js', '.spec.ts', '.test.js', '.test.ts', '.e2e.ts', '.input.js'],
    selenium: ['.test.js', '.spec.js', '.test.ts'],
    'selenium-java': ['.java'],
    'selenium-python': ['.py'],
    puppeteer: ['.spec.js', '.spec.ts', '.test.js', '.test.ts', '.input.js'],
    testcafe: ['.test.js', '.test.ts', '.spec.js'],
    webdriverio: ['.test.js', '.test.ts', '.spec.js', '.spec.ts', '.e2e.js', '.e2e.ts', '.input.js'],
    junit4: ['.java'],
    junit5: ['.java', '.input.java'],
    testng: ['.java', '.input.java'],
    pytest: ['.py', '.input.py'],
    unittest: ['.py', '.input.py'],
    nose2: ['.py', '.input.py'],
  };
  return map[framework] || ['.test.js', '.spec.js', '.input.js'];
}

// Calculate line-level fidelity between original and round-tripped output
function calculateFidelity(original, roundtripped) {
  const origLines = original.split('\n');
  const rtLines = roundtripped.split('\n');

  // LCS-based similarity
  let matches = 0;
  const rtSet = new Set(rtLines.map((l) => l.trim()));
  for (const line of origLines) {
    if (rtSet.has(line.trim()) && line.trim().length > 0) {
      matches++;
    }
  }

  const maxLines = Math.max(origLines.length, rtLines.length);
  if (maxLines === 0) return 1.0;

  return matches / maxLines;
}

// Calculate structural fidelity (describe/it/test blocks preserved)
function calculateStructuralFidelity(original, roundtripped) {
  const structurePatterns = [
    /\b(describe|it|test|beforeEach|afterEach|beforeAll|afterAll)\s*\(/g,
    /\b(expect)\s*\(/g,
    /@(Test|Before|After|BeforeEach|AfterEach|BeforeAll|AfterAll)\b/g,
    /\b(def test_\w+|class \w+.*TestCase|assert\s)/g,
  ];

  let origCount = 0;
  let rtCount = 0;

  for (const pat of structurePatterns) {
    origCount += (original.match(pat) || []).length;
    pat.lastIndex = 0;
    rtCount += (roundtripped.match(pat) || []).length;
    pat.lastIndex = 0;
  }

  if (origCount === 0 && rtCount === 0) return 1.0;
  if (origCount === 0) return 0;

  return Math.min(rtCount / origCount, 1.0);
}

// Main
async function main() {
  console.log('=== Hamlet Round-Trip Benchmark ===');
  console.log(`Direction: ${from} -> ${to}${forwardOnly ? '' : ` -> ${from}`}`);
  console.log(`Mode: ${forwardOnly ? 'forward-only' : 'round-trip'}`);
  console.log(`Input: ${inputDir}`);
  console.log(`Pattern: ${pattern}`);
  console.log('');

  const files = (await findFiles(inputDir, pattern)).slice(0, maxFiles);
  console.log(`Found ${files.length} test files`);
  console.log('');

  const results = {
    direction: `${from}-${to}-${from}`,
    inputDir,
    totalFiles: files.length,
    forwardOk: 0,
    forwardFail: 0,
    roundtripOk: 0,
    roundtripFail: 0,
    files: [],
    todoCount: 0,
  };

  let forwardConverter, reverseConverter;
  try {
    forwardConverter = await ConverterFactory.createConverter(from, to);
  } catch (e) {
    console.error(`Failed to create forward converter: ${e.message}`);
    process.exit(1);
  }

  if (!forwardOnly) {
    try {
      reverseConverter = await ConverterFactory.createConverter(to, from);
    } catch (e) {
      console.error(`Failed to create reverse converter: ${e.message}`);
      console.error('Use --forward-only to run forward conversion only.');
      process.exit(1);
    }
  }

  for (const file of files) {
    const relPath = path.relative(inputDir, file);
    const original = await fs.readFile(file, 'utf8');

    try {
      // Forward: from -> to
      const converted = await forwardConverter.convert(original);
      if (!converted || converted.length === 0) {
        results.forwardFail++;
        console.log(`  [FWD-EMPTY]  ${relPath}`);
        results.files.push({
          file: relPath,
          status: 'forward-empty',
          fidelity: 0,
          structuralFidelity: 0,
        });
        continue;
      }
      results.forwardOk++;

      // Count HAMLET-TODOs
      const todos = (converted.match(/HAMLET-TODO/g) || []).length;
      results.todoCount += todos;

      if (forwardOnly) {
        // Forward-only mode: no round-trip, just report forward success
        console.log(
          `  [FWD-OK]  ${relPath}  todos=${todos}  (${original.split('\n').length}L -> ${converted.split('\n').length}L)`,
        );
        results.files.push({
          file: relPath,
          status: 'forward-ok',
          origLines: original.split('\n').length,
          convertedLines: converted.split('\n').length,
          todos,
        });
        continue;
      }

      // Round-trip: to -> from
      try {
        const roundtripped = await reverseConverter.convert(converted);
        if (!roundtripped || roundtripped.length === 0) {
          results.roundtripFail++;
          console.log(`  [RT-EMPTY]   ${relPath}`);
          results.files.push({
            file: relPath,
            status: 'roundtrip-empty',
            fidelity: 0,
            structuralFidelity: 0,
            todos,
          });
          continue;
        }
        results.roundtripOk++;

        const fidelity = calculateFidelity(original, roundtripped);
        const structFidelity = calculateStructuralFidelity(
          original,
          roundtripped,
        );

        const status = fidelity > 0.8 ? 'OK' : fidelity > 0.5 ? 'FAIR' : 'LOW';
        console.log(
          `  [${status.padEnd(4)}]  ${relPath}  fidelity=${fidelity.toFixed(3)}  structural=${structFidelity.toFixed(3)}  todos=${todos}  (${original.split('\n').length}L -> ${converted.split('\n').length}L -> ${roundtripped.split('\n').length}L)`,
        );

        if (verbose && fidelity < 0.5) {
          console.log(`         Original (first 5 lines): ${original.split('\n').slice(0, 5).join(' | ')}`);
          console.log(`         Roundtrip (first 5 lines): ${roundtripped.split('\n').slice(0, 5).join(' | ')}`);
        }

        results.files.push({
          file: relPath,
          status: 'ok',
          fidelity,
          structuralFidelity: structFidelity,
          origLines: original.split('\n').length,
          convertedLines: converted.split('\n').length,
          roundtripLines: roundtripped.split('\n').length,
          todos,
        });
      } catch (e) {
        results.roundtripFail++;
        console.log(`  [RT-FAIL]    ${relPath}  ${e.message}`);
        results.files.push({
          file: relPath,
          status: 'roundtrip-error',
          error: e.message,
          todos,
        });
      }
    } catch (e) {
      results.forwardFail++;
      console.log(`  [FWD-FAIL]   ${relPath}  ${e.message}`);
      results.files.push({
        file: relPath,
        status: 'forward-error',
        error: e.message,
      });
    }
  }

  // Aggregate metrics
  const okFiles = results.files.filter((f) => f.status === 'ok');
  const avgFidelity =
    okFiles.length > 0
      ? okFiles.reduce((s, f) => s + f.fidelity, 0) / okFiles.length
      : 0;
  const avgStructural =
    okFiles.length > 0
      ? okFiles.reduce((s, f) => s + f.structuralFidelity, 0) / okFiles.length
      : 0;
  const medianFidelity = (() => {
    if (okFiles.length === 0) return 0;
    const sorted = okFiles.map((f) => f.fidelity).sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
  })();

  results.avgFidelity = avgFidelity;
  results.medianFidelity = medianFidelity;
  results.avgStructuralFidelity = avgStructural;

  console.log('');
  console.log('=== Summary ===');
  console.log(`Total files:            ${results.totalFiles}`);
  console.log(`Forward OK:             ${results.forwardOk}`);
  console.log(`Forward FAIL:           ${results.forwardFail}`);
  console.log(`Round-trip OK:          ${results.roundtripOk}`);
  console.log(`Round-trip FAIL:        ${results.roundtripFail}`);
  console.log(`Avg line fidelity:      ${avgFidelity.toFixed(4)}`);
  console.log(`Median line fidelity:   ${medianFidelity.toFixed(4)}`);
  console.log(`Avg structural fidelity: ${avgStructural.toFixed(4)}`);
  console.log(`Total HAMLET-TODOs:     ${results.todoCount}`);

  // Save results
  const resultsDir = path.join(__dirname, 'results');
  await fs.mkdir(resultsDir, { recursive: true });
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const resultFile = path.join(
    resultsDir,
    `${from}-${to}-roundtrip-${timestamp}.json`,
  );
  await fs.writeFile(resultFile, JSON.stringify(results, null, 2));
  console.log(`\nResults saved to: ${resultFile}`);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
