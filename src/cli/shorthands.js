/**
 * Shorthand command definitions for all supported conversion directions.
 *
 * Each direction gets two aliases:
 *   - Numeric form: jest2vt, cy2pw, etc.
 *   - Long form: jesttovt, cytopw, etc.
 */

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
 * All 25 supported conversion directions (from ConverterFactory.getSupportedConversions()).
 */
const DIRECTIONS = [
  // JavaScript E2E / Browser
  { from: 'cypress', to: 'playwright' },
  { from: 'cypress', to: 'selenium' },
  { from: 'playwright', to: 'cypress' },
  { from: 'playwright', to: 'selenium' },
  { from: 'selenium', to: 'cypress' },
  { from: 'selenium', to: 'playwright' },
  { from: 'cypress', to: 'webdriverio' },
  { from: 'webdriverio', to: 'cypress' },
  { from: 'webdriverio', to: 'playwright' },
  { from: 'playwright', to: 'webdriverio' },
  { from: 'puppeteer', to: 'playwright' },
  { from: 'playwright', to: 'puppeteer' },
  { from: 'testcafe', to: 'playwright' },
  { from: 'testcafe', to: 'cypress' },
  // JavaScript Unit Testing
  { from: 'jest', to: 'vitest' },
  { from: 'mocha', to: 'jest' },
  { from: 'jasmine', to: 'jest' },
  { from: 'jest', to: 'mocha' },
  { from: 'jest', to: 'jasmine' },
  // Java
  { from: 'junit4', to: 'junit5' },
  { from: 'junit5', to: 'testng' },
  { from: 'testng', to: 'junit5' },
  // Python
  { from: 'pytest', to: 'unittest' },
  { from: 'unittest', to: 'pytest' },
  { from: 'nose2', to: 'pytest' },
];

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
 */
export const CONVERSION_CATEGORIES = [
  {
    name: 'JavaScript E2E / Browser',
    directions: [
      {
        from: 'cypress',
        to: 'playwright',
        shorthands: buildAliasArray('cypress', 'playwright'),
      },
      {
        from: 'cypress',
        to: 'selenium',
        shorthands: buildAliasArray('cypress', 'selenium'),
      },
      {
        from: 'playwright',
        to: 'cypress',
        shorthands: buildAliasArray('playwright', 'cypress'),
      },
      {
        from: 'playwright',
        to: 'selenium',
        shorthands: buildAliasArray('playwright', 'selenium'),
      },
      {
        from: 'selenium',
        to: 'cypress',
        shorthands: buildAliasArray('selenium', 'cypress'),
      },
      {
        from: 'selenium',
        to: 'playwright',
        shorthands: buildAliasArray('selenium', 'playwright'),
      },
      {
        from: 'cypress',
        to: 'webdriverio',
        shorthands: buildAliasArray('cypress', 'webdriverio'),
      },
      {
        from: 'webdriverio',
        to: 'cypress',
        shorthands: buildAliasArray('webdriverio', 'cypress'),
      },
      {
        from: 'webdriverio',
        to: 'playwright',
        shorthands: buildAliasArray('webdriverio', 'playwright'),
      },
      {
        from: 'playwright',
        to: 'webdriverio',
        shorthands: buildAliasArray('playwright', 'webdriverio'),
      },
      {
        from: 'puppeteer',
        to: 'playwright',
        shorthands: buildAliasArray('puppeteer', 'playwright'),
      },
      {
        from: 'playwright',
        to: 'puppeteer',
        shorthands: buildAliasArray('playwright', 'puppeteer'),
      },
      {
        from: 'testcafe',
        to: 'playwright',
        shorthands: buildAliasArray('testcafe', 'playwright'),
      },
      {
        from: 'testcafe',
        to: 'cypress',
        shorthands: buildAliasArray('testcafe', 'cypress'),
      },
    ],
  },
  {
    name: 'JavaScript Unit Testing',
    directions: [
      {
        from: 'jest',
        to: 'vitest',
        shorthands: buildAliasArray('jest', 'vitest'),
      },
      {
        from: 'mocha',
        to: 'jest',
        shorthands: buildAliasArray('mocha', 'jest'),
      },
      {
        from: 'jasmine',
        to: 'jest',
        shorthands: buildAliasArray('jasmine', 'jest'),
      },
      {
        from: 'jest',
        to: 'mocha',
        shorthands: buildAliasArray('jest', 'mocha'),
      },
      {
        from: 'jest',
        to: 'jasmine',
        shorthands: buildAliasArray('jest', 'jasmine'),
      },
    ],
  },
  {
    name: 'Java',
    directions: [
      {
        from: 'junit4',
        to: 'junit5',
        shorthands: buildAliasArray('junit4', 'junit5'),
      },
      {
        from: 'junit5',
        to: 'testng',
        shorthands: buildAliasArray('junit5', 'testng'),
      },
      {
        from: 'testng',
        to: 'junit5',
        shorthands: buildAliasArray('testng', 'junit5'),
      },
    ],
  },
  {
    name: 'Python',
    directions: [
      {
        from: 'pytest',
        to: 'unittest',
        shorthands: buildAliasArray('pytest', 'unittest'),
      },
      {
        from: 'unittest',
        to: 'pytest',
        shorthands: buildAliasArray('unittest', 'pytest'),
      },
      {
        from: 'nose2',
        to: 'pytest',
        shorthands: buildAliasArray('nose2', 'pytest'),
      },
    ],
  },
];

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
 * Get the FRAMEWORK_ABBREV mapping (for doctor/debug commands).
 */
export { FRAMEWORK_ABBREV };
