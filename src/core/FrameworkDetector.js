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
      // Cypress-specific patterns
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
      // Playwright-specific patterns
      commands: [
        /page\./g,
        /browser\./g,
        /context\./g
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
        /\.spec\.(js|ts|jsx|tsx)$/,
        /\.test\.(js|ts|jsx|tsx)$/,
        /tests\//,
        /e2e\//
      ],
      keywords: [
        /page\.goto\(/,
        /page\.locator\(/,
        /page\.getBy/,
        /expect\([^)]+\)\.toBeVisible\(/,
        /expect\([^)]+\)\.toHaveText\(/,
        /test\(['"]/,
        /test\.describe\(/
      ]
    },

    [FRAMEWORKS.SELENIUM]: {
      // Selenium-specific patterns
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
        /wdio\.conf\.(js|ts)/,
        /selenium\.config\.(js|ts)/
      ],
      filePatterns: [
        /\.test\.(js|ts)$/,
        /\.spec\.(js|ts)$/,
        /test\//,
        /specs\//
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
      /\.should\(/
    ];

    return testPatterns.some(pattern => pattern.test(content));
  }
}
