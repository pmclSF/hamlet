import { FRAMEWORKS } from './ConverterFactory.js';

/**
 * Auto-detect the testing framework from test file content
 */
export class FrameworkDetector {
  /**
   * Framework detection patterns
   */
  static patterns = {
    [FRAMEWORKS.CYPRESS]: {
      commands: [
        /cy\./g,
        /cypress\./gi,
        /Cypress\./g
      ],
      imports: [
        /from\s+['"]cypress['"]/,
        /require\s*\(\s*['"]cypress['"]\s*\)/
      ],
      config: [
        /cypress\.config\.(js|ts|mjs)/,
        /cypress\.json/
      ],
      filePatterns: [
        /\.cy\.(js|ts|jsx|tsx)$/,
        /cypress\/e2e\//,
        /cypress\/integration\//,
        /cypress\/component\//
      ],
      keywords: [
        /cy\.visit\(/,
        /cy\.get\(/,
        /cy\.contains\(/,
        /cy\.intercept\(/,
        /cy\.request\(/,
        /\.should\(['"]be\./,
        /\.should\(['"]have\./
      ]
    },

    [FRAMEWORKS.PLAYWRIGHT]: {
      commands: [
        /page\.(goto|locator|getBy)/g,
        /test\.describe\(/g
      ],
      imports: [
        /from\s+['"]@playwright\/test['"]/,
        /from\s+['"]playwright['"]/,
        /require\s*\(\s*['"]@playwright\/test['"]\s*\)/,
        /require\s*\(\s*['"]playwright['"]\s*\)/
      ],
      config: [
        /playwright\.config\.(js|ts|mjs)/
      ],
      filePatterns: [
        /\.spec\.(js|ts|jsx|tsx)$/
      ],
      keywords: [
        /page\.goto\(/,
        /page\.locator\(/,
        /page\.getBy/,
        /expect\([^)]+\)\.toBeVisible\(/,
        /expect\([^)]+\)\.toHaveText\(/,
        /test\.describe\(/,
        /test\.beforeEach\(/
      ]
    },

    [FRAMEWORKS.SELENIUM]: {
      commands: [
        /driver\./g,
        /webdriver\./gi,
        /WebDriver\./g
      ],
      imports: [
        /from\s+['"]selenium-webdriver['"]/,
        /require\s*\(\s*['"]selenium-webdriver['"]\s*\)/,
        /from\s+['"]webdriver['"]/
      ],
      config: [
        /selenium\.config\.(js|ts)/
      ],
      filePatterns: [
        /selenium/i
      ],
      keywords: [
        /driver\.get\(/,
        /driver\.findElement\(/,
        /driver\.findElements\(/,
        /By\.(css|xpath|id|name|className)\(/,
        /\.sendKeys\(/,
        /driver\.wait\(/,
        /until\.elementLocated\(/
      ]
    },

    [FRAMEWORKS.JEST]: {
      commands: [
        /jest\./g
      ],
      imports: [
        /from\s+['"]@jest\/globals['"]/,
        /from\s+['"]jest['"]/
      ],
      config: [
        /jest\.config\.(js|ts|mjs|cjs)/
      ],
      filePatterns: [
        /\.test\.(js|ts|jsx|tsx)$/,
        /__tests__\//
      ],
      keywords: [
        /jest\.fn\(/,
        /jest\.mock\(/,
        /jest\.spyOn\(/,
        /jest\.clearAllMocks/,
        /jest\.useFakeTimers/,
        /test\.each/,
        /expect\([^)]*\)\.(toBe|toEqual|toContain|toThrow|toHaveBeenCalled)/
      ]
    },

    [FRAMEWORKS.VITEST]: {
      commands: [
        /vi\./g
      ],
      imports: [
        /from\s+['"]vitest['"]/
      ],
      config: [
        /vitest\.config\.(js|ts|mjs)/,
        /vite\.config\.(js|ts|mjs)/
      ],
      filePatterns: [
        /\.test\.(js|ts|jsx|tsx)$/,
        /\.spec\.(js|ts|jsx|tsx)$/
      ],
      keywords: [
        /vi\.fn\(/,
        /vi\.mock\(/,
        /vi\.spyOn\(/,
        /vi\.useFakeTimers/,
        /vi\.useRealTimers/
      ]
    },

    [FRAMEWORKS.MOCHA]: {
      commands: [],
      imports: [
        /from\s+['"]mocha['"]/,
        /require\s*\(\s*['"]mocha['"]\s*\)/,
        /from\s+['"]chai['"]/,
        /require\s*\(\s*['"]chai['"]\s*\)/
      ],
      config: [
        /\.mocharc\.(yml|yaml|json|js|cjs)/
      ],
      filePatterns: [
        /\.test\.(js|ts)$/,
        /\.spec\.(js|ts)$/,
        /test\//
      ],
      keywords: [
        /\bcontext\s*\(/,
        /\bspecify\s*\(/,
        /\bsuiteSetup\s*\(/,
        /\bsuiteTeardown\s*\(/,
        /\bsetup\s*\(/,
        /\bteardown\s*\(/,
        /expect\([^)]*\)\.to\.(be|have|equal|deep|include)/,
        /assert\.(equal|deepEqual|strictEqual|ok|throws)/
      ]
    },

    [FRAMEWORKS.JASMINE]: {
      commands: [
        /jasmine\./g
      ],
      imports: [
        /from\s+['"]jasmine['"]/,
        /require\s*\(\s*['"]jasmine['"]\s*\)/
      ],
      config: [
        /jasmine\.json/,
        /jasmine\.config/
      ],
      filePatterns: [
        /\.spec\.(js|ts)$/,
        /spec\//
      ],
      keywords: [
        /jasmine\.createSpy/,
        /jasmine\.createSpyObj/,
        /\.and\.returnValue\(/,
        /\.and\.callThrough\(/,
        /\.and\.callFake\(/,
        /\bfdescribe\s*\(/,
        /\bfit\s*\(/,
        /\bxdescribe\s*\(/,
        /\bxit\s*\(/
      ]
    },

    [FRAMEWORKS.WEBDRIVERIO]: {
      commands: [
        /browser\.(url|getUrl|pause|keys)\(/g,
        /\$\(/g,
        /\$\$\(/g
      ],
      imports: [
        /from\s+['"]webdriverio['"]/,
        /from\s+['"]@wdio\//,
        /require\s*\(\s*['"]webdriverio['"]\s*\)/
      ],
      config: [
        /wdio\.conf\.(js|ts|mjs)/
      ],
      filePatterns: [
        /\.test\.(js|ts)$/,
        /\.spec\.(js|ts)$/,
        /wdio/i
      ],
      keywords: [
        /browser\.url\(/,
        /\$\(['"][^'"]+['"]\)/,
        /\$\$\(['"][^'"]+['"]\)/,
        /\.waitForDisplayed\(/,
        /\.waitForExist\(/,
        /\.setValue\(/,
        /\.getValue\(/,
        /\.isDisplayed\(/
      ]
    },

    [FRAMEWORKS.PUPPETEER]: {
      commands: [
        /puppeteer\./g
      ],
      imports: [
        /from\s+['"]puppeteer['"]/,
        /require\s*\(\s*['"]puppeteer['"]\s*\)/
      ],
      config: [],
      filePatterns: [
        /puppeteer/i
      ],
      keywords: [
        /puppeteer\.launch\(/,
        /browser\.newPage\(/,
        /page\.type\(/,
        /page\.\$eval\(/,
        /page\.\$\$eval\(/,
        /page\.waitForSelector\(/,
        /page\.evaluate\(/,
        /page\.screenshot\(/
      ]
    },

    [FRAMEWORKS.TESTCAFE]: {
      commands: [
        /\bt\./g
      ],
      imports: [
        /from\s+['"]testcafe['"]/,
        /require\s*\(\s*['"]testcafe['"]\s*\)/
      ],
      config: [
        /\.testcaferc\.(json|js|cjs)/
      ],
      filePatterns: [
        /testcafe/i
      ],
      keywords: [
        /\bfixture\s*\(/,
        /\bSelector\s*\(/,
        /\bClientFunction\s*\(/,
        /t\.typeText\(/,
        /t\.click\(/,
        /t\.expect\(/,
        /t\.navigateTo\(/
      ]
    },

    [FRAMEWORKS.JUNIT4]: {
      commands: [],
      imports: [
        /import\s+org\.junit\.Test/,
        /import\s+org\.junit\.Before/,
        /import\s+org\.junit\.After/,
        /import\s+org\.junit\.Assert/,
        /import\s+static\s+org\.junit\.Assert\.\*/,
        /import\s+org\.junit\.runner/
      ],
      config: [],
      filePatterns: [
        /Test\.java$/,
        /Tests\.java$/
      ],
      keywords: [
        /@Test\b/,
        /@Before\b(?!Each|All)/,
        /@After\b(?!Each|All)/,
        /@RunWith\(/,
        /@Rule\b/,
        /Assert\.(assertEquals|assertTrue|assertFalse|assertNull|assertNotNull)/,
        /@Test\s*\(\s*expected\s*=/
      ]
    },

    [FRAMEWORKS.JUNIT5]: {
      commands: [],
      imports: [
        /import\s+org\.junit\.jupiter/,
        /import\s+static\s+org\.junit\.jupiter\.api\.Assertions\.\*/
      ],
      config: [],
      filePatterns: [
        /Test\.java$/,
        /Tests\.java$/
      ],
      keywords: [
        /@Test\b/,
        /@BeforeEach\b/,
        /@AfterEach\b/,
        /@BeforeAll\b/,
        /@AfterAll\b/,
        /@DisplayName\(/,
        /@Nested\b/,
        /@ParameterizedTest\b/,
        /@CsvSource\(/,
        /@ValueSource\(/,
        /Assertions\.(assertEquals|assertTrue|assertFalse|assertThrows|assertAll)/,
        /assertThrows\(/
      ]
    },

    [FRAMEWORKS.TESTNG]: {
      commands: [],
      imports: [
        /import\s+org\.testng/,
        /import\s+static\s+org\.testng\.Assert\.\*/
      ],
      config: [
        /testng\.xml/
      ],
      filePatterns: [
        /Test\.java$/
      ],
      keywords: [
        /@Test\b/,
        /@BeforeMethod\b/,
        /@AfterMethod\b/,
        /@BeforeClass\b/,
        /@AfterClass\b/,
        /@DataProvider\b/,
        /@Test\s*\(\s*dataProvider\s*=/,
        /@Test\s*\(\s*expectedExceptions\s*=/,
        /Assert\.(assertEquals|assertTrue|assertFalse|assertNull|assertNotNull)/
      ]
    },

    [FRAMEWORKS.PYTEST]: {
      commands: [
        /pytest\./g
      ],
      imports: [
        /import\s+pytest/,
        /from\s+pytest\s+import/
      ],
      config: [
        /pytest\.ini/,
        /pyproject\.toml/,
        /conftest\.py/,
        /setup\.cfg/
      ],
      filePatterns: [
        /test_.*\.py$/,
        /.*_test\.py$/,
        /conftest\.py$/
      ],
      keywords: [
        /@pytest\.fixture/,
        /@pytest\.mark\./,
        /@pytest\.mark\.parametrize/,
        /pytest\.raises\(/,
        /pytest\.skip\(/,
        /\bdef test_\w+/,
        /\bassert\s+/
      ]
    },

    [FRAMEWORKS.UNITTEST]: {
      commands: [
        /unittest\./g
      ],
      imports: [
        /import\s+unittest/,
        /from\s+unittest\s+import/
      ],
      config: [],
      filePatterns: [
        /test_.*\.py$/,
        /.*_test\.py$/
      ],
      keywords: [
        /class\s+\w+\(unittest\.TestCase\)/,
        /self\.assert(Equal|True|False|Raises|In|NotIn|Is|IsNone|IsNotNone)/,
        /self\.setUp\(/,
        /self\.tearDown\(/,
        /unittest\.main\(/,
        /self\.subTest\(/,
        /unittest\.skip\(/
      ]
    },

    [FRAMEWORKS.NOSE2]: {
      commands: [
        /nose2\./g
      ],
      imports: [
        /import\s+nose2/,
        /from\s+nose2/,
        /from\s+nose2\.tools\s+import/
      ],
      config: [
        /nose2\.cfg/,
        /setup\.cfg/
      ],
      filePatterns: [
        /test_.*\.py$/
      ],
      keywords: [
        /@params\b/,
        /nose2\.tools/,
        /nose2\.discover/
      ]
    }
  };

  /**
   * Detect framework from file content
   * @param {string} content - Test file content
   * @returns {Object} - { framework: string, confidence: number, details: Object }
   */
  static detectFromContent(content) {
    const scores = {};
    const details = {};

    for (const [framework, patterns] of Object.entries(this.patterns)) {
      let score = 0;
      const matches = {
        commands: 0,
        imports: 0,
        keywords: 0
      };

      // Check commands (weight: 2)
      for (const pattern of patterns.commands) {
        const cmdMatches = content.match(pattern);
        if (cmdMatches) {
          score += cmdMatches.length * 2;
          matches.commands += cmdMatches.length;
        }
      }

      // Check imports (weight: 10 - strong indicator)
      for (const pattern of patterns.imports) {
        if (pattern.test(content)) {
          score += 10;
          matches.imports++;
        }
      }

      // Check keywords (weight: 3)
      for (const pattern of patterns.keywords) {
        if (pattern.test(content)) {
          score += 3;
          matches.keywords++;
        }
      }

      scores[framework] = score;
      details[framework] = matches;
    }

    // Find the framework with highest score
    let detectedFramework = null;
    let maxScore = 0;

    for (const [framework, score] of Object.entries(scores)) {
      if (score > maxScore) {
        maxScore = score;
        detectedFramework = framework;
      }
    }

    // Calculate confidence (0-1)
    const totalScore = Object.values(scores).reduce((a, b) => a + b, 0);
    const confidence = totalScore > 0 ? maxScore / totalScore : 0;

    return {
      framework: detectedFramework,
      confidence: Math.round(confidence * 100) / 100,
      scores,
      details
    };
  }

  /**
   * Detect framework from file path
   * @param {string} filePath - Path to test file
   * @returns {Object} - { framework: string, confidence: number }
   */
  static detectFromPath(filePath) {
    for (const [framework, patterns] of Object.entries(this.patterns)) {
      for (const pattern of patterns.filePatterns) {
        if (pattern.test(filePath)) {
          return {
            framework,
            confidence: 0.7, // Path-based detection is less certain
            reason: `Matched file pattern: ${pattern.toString()}`
          };
        }
      }

      for (const pattern of patterns.config) {
        if (pattern.test(filePath)) {
          return {
            framework,
            confidence: 0.9, // Config file is a strong indicator
            reason: `Matched config pattern: ${pattern.toString()}`
          };
        }
      }
    }

    return {
      framework: null,
      confidence: 0,
      reason: 'No matching patterns found'
    };
  }

  /**
   * Detect framework using both content and path
   * @param {string} content - File content
   * @param {string} filePath - File path
   * @returns {Object} - Combined detection result
   */
  static detect(content, filePath = '') {
    const contentResult = this.detectFromContent(content);
    const pathResult = this.detectFromPath(filePath);

    // If both agree, increase confidence
    if (contentResult.framework === pathResult.framework && contentResult.framework) {
      return {
        framework: contentResult.framework,
        confidence: Math.min(1, contentResult.confidence + 0.2),
        method: 'combined',
        contentAnalysis: contentResult,
        pathAnalysis: pathResult
      };
    }

    // If content detection is more confident, use it
    if (contentResult.confidence >= pathResult.confidence) {
      return {
        framework: contentResult.framework,
        confidence: contentResult.confidence,
        method: 'content',
        contentAnalysis: contentResult,
        pathAnalysis: pathResult
      };
    }

    // Otherwise use path detection
    return {
      framework: pathResult.framework,
      confidence: pathResult.confidence,
      method: 'path',
      contentAnalysis: contentResult,
      pathAnalysis: pathResult
    };
  }

  /**
   * Get all detectable frameworks
   * @returns {string[]}
   */
  static getDetectableFrameworks() {
    return Object.keys(this.patterns);
  }

  /**
   * Check if content appears to be a test file
   * @param {string} content - File content
   * @returns {boolean}
   */
  static isTestFile(content) {
    const testPatterns = [
      /describe\s*\(/,
      /it\s*\(/,
      /test\s*\(/,
      /expect\s*\(/,
      /assert\./,
      /\.should\(/,
      /@Test\b/,
      /def test_\w+/,
      /class\s+\w+\(unittest\.TestCase\)/,
      /fixture\s*\(/
    ];

    return testPatterns.some(pattern => pattern.test(content));
  }
}
