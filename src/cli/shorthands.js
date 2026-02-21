/**
 * Shorthand command definitions for all supported conversion directions.
 *
 * Each direction gets two aliases:
 *   - Numeric form: jest2vt, cy2pw, etc.
 *   - Long form: jesttovt, cytopw, etc.
 *
 * DIRECTIONS is generated from ConverterFactory.getSupportedConversions()
 * so it cannot drift from the actual supported set.
 */

import { ConverterFactory } from '../core/ConverterFactory.js';

/**
 * Abbreviations for framework names used in shorthand commands.
 */
const FRAMEWORK_ABBREV = {
  cypress: 'cy',
  playwright: 'pw',
  selenium: 'sel',
  jest: 'jest',
  vitest: 'vt',
  mocha: 'mocha',
  jasmine: 'jas',
  junit4: 'ju4',
  junit5: 'ju5',
  testng: 'tng',
  pytest: 'pyt',
  unittest: 'ut',
  nose2: 'nose2',
  webdriverio: 'wdio',
  puppeteer: 'pptr',
  testcafe: 'tcafe',
};

/**
 * All supported conversion directions — derived from ConverterFactory.
 */
const DIRECTIONS = ConverterFactory.getSupportedConversions().map((key) => {
  const [from, to] = key.split('-');
  return { from, to };
});

/**
 * Language category for a framework (for grouping in `list` output).
 */
const FRAMEWORK_CATEGORY = {
  cypress: 'JavaScript E2E / Browser',
  playwright: 'JavaScript E2E / Browser',
  selenium: 'JavaScript E2E / Browser',
  webdriverio: 'JavaScript E2E / Browser',
  puppeteer: 'JavaScript E2E / Browser',
  testcafe: 'JavaScript E2E / Browser',
  jest: 'JavaScript Unit Testing',
  vitest: 'JavaScript Unit Testing',
  mocha: 'JavaScript Unit Testing',
  jasmine: 'JavaScript Unit Testing',
  junit4: 'Java',
  junit5: 'Java',
  testng: 'Java',
  pytest: 'Python',
  unittest: 'Python',
  nose2: 'Python',
};

/**
 * Build shorthand aliases for a direction.
 * @param {string} from
 * @param {string} to
 * @returns {{ numeric: string, long: string }}
 */
function buildAliases(from, to) {
  const abbrevFrom = FRAMEWORK_ABBREV[from];
  const abbrevTo = FRAMEWORK_ABBREV[to];
  return {
    numeric: `${abbrevFrom}2${abbrevTo}`,
    long: `${abbrevFrom}to${abbrevTo}`,
  };
}

/**
 * Build alias array for a direction.
 * @param {string} from
 * @param {string} to
 * @returns {string[]}
 */
function buildAliasArray(from, to) {
  const { numeric, long } = buildAliases(from, to);
  const result = [numeric];
  if (long !== numeric) {
    result.push(long);
  }
  return result;
}

/**
 * Map of shorthand alias → { from, to }.
 * Each direction has two aliases (numeric "2" and long "to" form).
 */
export const SHORTHANDS = {};

for (const { from, to } of DIRECTIONS) {
  const { numeric, long } = buildAliases(from, to);
  SHORTHANDS[numeric] = { from, to };
  // Only add long form if it differs from numeric
  if (long !== numeric) {
    SHORTHANDS[long] = { from, to };
  }
}

/**
 * Categorized conversion directions for the `list` command.
 * Generated from DIRECTIONS grouped by source framework category.
 */
export const CONVERSION_CATEGORIES = (() => {
  const categoryMap = new Map();

  for (const { from, to } of DIRECTIONS) {
    const categoryName = FRAMEWORK_CATEGORY[from] || 'Other';
    if (!categoryMap.has(categoryName)) {
      categoryMap.set(categoryName, []);
    }
    categoryMap.get(categoryName).push({
      from,
      to,
      shorthands: buildAliasArray(from, to),
    });
  }

  // Stable ordering: JS E2E, JS Unit, Java, Python
  const ORDER = [
    'JavaScript E2E / Browser',
    'JavaScript Unit Testing',
    'Java',
    'Python',
  ];

  return ORDER.filter((name) => categoryMap.has(name)).map((name) => ({
    name,
    directions: categoryMap.get(name),
  }));
})();

/**
 * Get the FRAMEWORK_ABBREV mapping (for doctor/debug commands).
 */
export { FRAMEWORK_ABBREV };
