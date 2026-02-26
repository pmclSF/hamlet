/**
 * Baseline regression tests for round-trip fidelity.
 *
 * Reads baselines.json and verifies that current fidelity scores
 * do not drop below the recorded baseline (with a small epsilon
 * to prevent noise from blocking CI).
 */

import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import {
  convert,
  calculateFidelity,
  calculateStructuralFidelity,
} from './roundtrip.helper.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const fixturesDir = path.resolve(__dirname, 'fixtures');
const baselinesPath = path.resolve(__dirname, 'baselines.json');

// Epsilon: fidelity may fluctuate slightly due to whitespace / ordering
const EPSILON = 0.01;
// Threshold above which we log a suggestion to update baselines
const IMPROVEMENT_THRESHOLD = 0.02;

// Map pair keys back to conversion metadata
const PAIR_META = {
  'jest-mocha': { from: 'jest', to: 'mocha', ext: 'js' },
  'jest-jasmine': { from: 'jest', to: 'jasmine', ext: 'js' },
  'cypress-playwright': { from: 'cypress', to: 'playwright', ext: 'js' },
  'cypress-webdriverio': { from: 'cypress', to: 'webdriverio', ext: 'js' },
  'playwright-webdriverio': {
    from: 'playwright',
    to: 'webdriverio',
    ext: 'js',
  },
  'playwright-puppeteer': { from: 'playwright', to: 'puppeteer', ext: 'js' },
  'junit5-testng': { from: 'junit5', to: 'testng', ext: 'java' },
  'pytest-unittest': { from: 'pytest', to: 'unittest', ext: 'py' },
  'cypress-selenium': { from: 'cypress', to: 'selenium', ext: 'js' },
  'playwright-selenium': { from: 'playwright', to: 'selenium', ext: 'js' },
};

// Map pair keys to fixture directory names
const PAIR_DIR = {
  'jest-mocha': 'jest-mocha',
  'jest-jasmine': 'jest-jasmine',
  'cypress-playwright': 'cypress-playwright',
  'cypress-webdriverio': 'cypress-wdio',
  'playwright-webdriverio': 'playwright-wdio',
  'playwright-puppeteer': 'playwright-puppeteer',
  'junit5-testng': 'junit5-testng',
  'pytest-unittest': 'pytest-unittest',
  'cypress-selenium': 'cypress-selenium',
  'playwright-selenium': 'playwright-selenium',
};

let baselines;

beforeAll(async () => {
  try {
    const raw = await fs.readFile(baselinesPath, 'utf8');
    baselines = JSON.parse(raw);
  } catch (_e) {
    // baselines.json doesn't exist yet — skip all tests gracefully
    baselines = null;
  }
});

describe('Baseline Regression', () => {
  it('should have a valid baselines.json file', () => {
    if (!baselines) {
      console.warn(
        'baselines.json not found — run `node scripts/update-baselines.js` to generate'
      );
      return;
    }
    expect(baselines.version).toBe(1);
    expect(baselines.pairs).toBeDefined();
    expect(Object.keys(baselines.pairs).length).toBeGreaterThan(0);
  });

  // Dynamically generate tests from all possible pairs
  for (const [pairKey, meta] of Object.entries(PAIR_META)) {
    const dir = PAIR_DIR[pairKey];

    for (const complexity of ['simple', 'medium', 'complex']) {
      it(`${pairKey} ${complexity}: fidelity >= baseline`, async () => {
        if (!baselines) {
          console.warn('No baselines.json — skipping');
          return;
        }

        const pairBaseline = baselines.pairs[pairKey];
        if (!pairBaseline || !pairBaseline[complexity]) {
          // No baseline recorded for this pair/complexity — skip
          return;
        }

        const fixturePath = path.join(
          fixturesDir,
          dir,
          `${complexity}.input.${meta.ext}`
        );

        let original;
        try {
          original = await fs.readFile(fixturePath, 'utf8');
        } catch (_e) {
          console.warn(`Fixture not found: ${fixturePath}`);
          return;
        }

        // Forward: from → to
        const forward = await convert(original, meta.from, meta.to);
        expect(forward.code).toBeTruthy();

        // Reverse: to → from
        const reverse = await convert(forward.code, meta.to, meta.from);
        expect(reverse.code).toBeTruthy();

        const lineFidelity = calculateFidelity(original, reverse.code);
        const structuralFidelity = calculateStructuralFidelity(
          original,
          reverse.code
        );

        const baselineLine = pairBaseline[complexity].lineFidelity;
        const baselineStructural = pairBaseline[complexity].structuralFidelity;

        // Assert no regression (with epsilon)
        expect(lineFidelity).toBeGreaterThanOrEqual(baselineLine - EPSILON);
        expect(structuralFidelity).toBeGreaterThanOrEqual(
          baselineStructural - EPSILON
        );

        // Log when fidelity has improved significantly
        if (lineFidelity > baselineLine + IMPROVEMENT_THRESHOLD) {
          console.log(
            `  📈 ${pairKey}/${complexity} line fidelity improved: ` +
              `${baselineLine.toFixed(3)} → ${lineFidelity.toFixed(3)} ` +
              `(consider running update-baselines.js)`
          );
        }
        if (structuralFidelity > baselineStructural + IMPROVEMENT_THRESHOLD) {
          console.log(
            `  📈 ${pairKey}/${complexity} structural fidelity improved: ` +
              `${baselineStructural.toFixed(3)} → ${structuralFidelity.toFixed(3)} ` +
              `(consider running update-baselines.js)`
          );
        }
      });
    }
  }
});
