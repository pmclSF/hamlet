#!/usr/bin/env node
/**
 * Generate round-trip fidelity baselines.
 *
 * Iterates all bidirectional pairs × 3 complexity levels, runs A→B→A
 * conversion, computes line-level and structural fidelity, and writes
 * test/integration/roundtrip/baselines.json.
 *
 * Usage:
 *   node scripts/update-baselines.js            # write baselines.json
 *   node scripts/update-baselines.js --dry-run   # print without writing
 */

import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '..');
const fixturesDir = path.resolve(
  rootDir,
  'test/integration/roundtrip/fixtures'
);
const baselinesPath = path.resolve(
  rootDir,
  'test/integration/roundtrip/baselines.json'
);

const { ConverterFactory } = await import(
  path.join(rootDir, 'src/core/ConverterFactory.js')
);

// ── Configuration ────────────────────────────────────────────────────

const ROUND_TRIP_PAIRS = [
  { dir: 'jest-mocha', from: 'jest', to: 'mocha', ext: 'js' },
  { dir: 'jest-jasmine', from: 'jest', to: 'jasmine', ext: 'js' },
  { dir: 'cypress-playwright', from: 'cypress', to: 'playwright', ext: 'js' },
  { dir: 'cypress-wdio', from: 'cypress', to: 'webdriverio', ext: 'js' },
  { dir: 'playwright-wdio', from: 'playwright', to: 'webdriverio', ext: 'js' },
  {
    dir: 'playwright-puppeteer',
    from: 'playwright',
    to: 'puppeteer',
    ext: 'js',
  },
  { dir: 'junit5-testng', from: 'junit5', to: 'testng', ext: 'java' },
  { dir: 'pytest-unittest', from: 'pytest', to: 'unittest', ext: 'py' },
  { dir: 'cypress-selenium', from: 'cypress', to: 'selenium', ext: 'js' },
  {
    dir: 'playwright-selenium',
    from: 'playwright',
    to: 'selenium',
    ext: 'js',
  },
];

const COMPLEXITY_LEVELS = ['simple', 'medium', 'complex'];

// ── Fidelity calculations (same as roundtrip.helper.js) ──────────────

function calculateFidelity(original, roundtripped) {
  const origLines = original.split('\n');
  const rtLines = roundtripped.split('\n');

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

// ── Main ─────────────────────────────────────────────────────────────

const dryRun = process.argv.includes('--dry-run');

async function main() {
  console.log('=== Hamlet Baseline Generator ===');
  console.log(`Mode: ${dryRun ? 'dry-run (no write)' : 'write'}`);
  console.log('');

  const baselines = {
    version: 1,
    generatedAt: new Date().toISOString(),
    pairs: {},
  };

  let totalPairs = 0;
  let skippedPairs = 0;

  for (const pair of ROUND_TRIP_PAIRS) {
    const pairKey = `${pair.from}-${pair.to}`;
    const pairBaselines = {};

    let forwardConverter, reverseConverter;
    try {
      forwardConverter = await ConverterFactory.createConverter(
        pair.from,
        pair.to
      );
      reverseConverter = await ConverterFactory.createConverter(
        pair.to,
        pair.from
      );
    } catch (e) {
      console.warn(`  Skipping ${pairKey}: ${e.message}`);
      skippedPairs += COMPLEXITY_LEVELS.length;
      continue;
    }

    for (const complexity of COMPLEXITY_LEVELS) {
      const fixturePath = path.join(
        fixturesDir,
        pair.dir,
        `${complexity}.input.${pair.ext}`
      );

      totalPairs++;

      let original;
      try {
        original = await fs.readFile(fixturePath, 'utf8');
      } catch (_e) {
        console.warn(`  Skipping ${pairKey}/${complexity}: fixture not found`);
        skippedPairs++;
        continue;
      }

      try {
        // Forward: from → to
        const converted = await forwardConverter.convert(original);
        if (!converted || converted.length === 0) {
          console.warn(
            `  Skipping ${pairKey}/${complexity}: forward conversion empty`
          );
          skippedPairs++;
          continue;
        }

        // Reverse: to → from
        const roundtripped = await reverseConverter.convert(converted);
        if (!roundtripped || roundtripped.length === 0) {
          console.warn(
            `  Skipping ${pairKey}/${complexity}: reverse conversion empty`
          );
          skippedPairs++;
          continue;
        }

        const lineFidelity = calculateFidelity(original, roundtripped);
        const structuralFidelity = calculateStructuralFidelity(
          original,
          roundtripped
        );

        pairBaselines[complexity] = {
          lineFidelity: Math.round(lineFidelity * 1000) / 1000,
          structuralFidelity: Math.round(structuralFidelity * 1000) / 1000,
        };

        const status =
          lineFidelity > 0.8 ? 'OK' : lineFidelity > 0.5 ? 'FAIR' : 'LOW';
        console.log(
          `  [${status.padEnd(4)}] ${pairKey}/${complexity}  ` +
            `line=${lineFidelity.toFixed(3)}  structural=${structuralFidelity.toFixed(3)}`
        );
      } catch (e) {
        console.warn(
          `  Skipping ${pairKey}/${complexity}: conversion error: ${e.message}`
        );
        skippedPairs++;
      }
    }

    if (Object.keys(pairBaselines).length > 0) {
      baselines.pairs[pairKey] = pairBaselines;
    }
  }

  console.log('');
  console.log(`Total: ${totalPairs} pair/complexity combos`);
  console.log(`Skipped: ${skippedPairs}`);
  console.log(`Baselined: ${totalPairs - skippedPairs}`);

  if (dryRun) {
    console.log('\n--- baselines.json (dry-run) ---');
    console.log(JSON.stringify(baselines, null, 2));
  } else {
    await fs.writeFile(
      baselinesPath,
      JSON.stringify(baselines, null, 2) + '\n'
    );
    console.log(`\nWritten to: ${baselinesPath}`);
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
