import fs from 'fs/promises';
import path from 'path';
import chalk from 'chalk';

let _PNG = null;
let _pixelmatch = null;

/**
 * Lazily load pngjs and pixelmatch. These are only needed when actually
 * comparing screenshots, not for name-matching or other lightweight operations.
 */
async function loadDeps() {
  if (_PNG && _pixelmatch) return;
  _PNG = (await import('pngjs')).PNG;
  _pixelmatch = (await import('pixelmatch')).default;
}

/**
 * Handles visual comparison between Cypress and Playwright tests
 */
export class VisualComparison {
  constructor(options = {}) {
    this.options = {
      threshold: options.threshold || 0.1,
      includeLogs: options.includeLogs ?? true,
      saveSnapshots: options.saveSnapshots ?? true,
      snapshotDir: options.snapshotDir || 'snapshots',
      ...options,
    };

    this.results = {
      comparisons: [],
      matches: 0,
      mismatches: 0,
      errors: [],
    };
  }

  /**
   * Compare screenshots between Cypress and Playwright tests
   * @param {string} cypressDir - Cypress project directory
   * @param {string} playwrightDir - Playwright project directory
   * @returns {Promise<Object>} - Comparison results
   */
  async compareProjects(cypressDir, playwrightDir) {
    try {
      console.log(chalk.blue('\nStarting visual comparison...'));

      // Ensure snapshot directory exists
      await fs.mkdir(this.options.snapshotDir, { recursive: true });

      // Find all screenshot files
      const cypressScreenshots = await this.findScreenshots(cypressDir);
      const playwrightScreenshots = await this.findScreenshots(playwrightDir);

      // Process each comparison
      for (const cypressShot of cypressScreenshots) {
        const playwrightShot = this.findMatchingScreenshot(
          cypressShot,
          playwrightScreenshots
        );

        if (playwrightShot) {
          await this.compareScreenshots(cypressShot, playwrightShot);
        } else {
          this.results.errors.push({
            type: 'missing',
            cypressShot,
            message: 'No matching Playwright screenshot found',
          });
        }
      }

      // Generate report
      const report = await this.generateReport();

      console.log(chalk.green('\n✓ Visual comparison completed'));
      this.logSummary();

      return report;
    } catch (error) {
      console.error(chalk.red('Error during visual comparison:'), error);
      throw error;
    }
  }

  /**
   * Find all screenshot files in a directory
   * @param {string} dir - Directory to search
   * @returns {Promise<string[]>} - Array of screenshot paths
   */
  async findScreenshots(dir) {
    const screenshots = [];

    async function scan(directory) {
      const entries = await fs.readdir(directory, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(directory, entry.name);

        if (entry.isDirectory()) {
          await scan(fullPath);
        } else if (/\.(png|jpg)$/.test(entry.name)) {
          screenshots.push(fullPath);
        }
      }
    }

    await scan(dir);
    return screenshots;
  }

  /**
   * Find matching screenshot from Playwright tests
   * @param {string} cypressShot - Cypress screenshot path
   * @param {string[]} playwrightShots - Playwright screenshot paths
   * @returns {string|null} - Matching screenshot path or null
   */
  findMatchingScreenshot(cypressShot, playwrightShots) {
    const cypressName = path.basename(cypressShot, path.extname(cypressShot));
    const normalizedCypress = this._normalizeName(cypressName);

    // Build index on first call for O(1) exact-match lookup
    const index = new Map();
    for (const shot of playwrightShots) {
      const name = path.basename(shot, path.extname(shot));
      index.set(this._normalizeName(name), shot);
    }

    // Try exact match first
    if (index.has(normalizedCypress)) {
      return index.get(normalizedCypress);
    }

    // Fall back to fuzzy matching with early-exit Levenshtein
    return playwrightShots.find((shot) => {
      const shotName = path.basename(shot, path.extname(shot));
      const norm = this._normalizeName(shotName);
      const maxLen = Math.max(normalizedCypress.length, norm.length);
      if (maxLen === 0) return true;
      const maxDist = Math.floor(maxLen * 0.2); // similarity > 0.8
      return (
        this.levenshteinDistance(normalizedCypress, norm, maxDist) <= maxDist
      );
    });
  }

  /**
   * Normalize a screenshot name by stripping framework prefixes/suffixes.
   */
  _normalizeName(name) {
    return name
      .replace(/(cypress|playwright)-?/, '')
      .replace(/-?screenshot/, '');
  }

  /**
   * Calculate similarity between two screenshot names
   * @param {string} name1 - First name
   * @param {string} name2 - Second name
   * @returns {number} - Similarity score (0-1)
   */
  calculateNameSimilarity(name1, name2) {
    // Remove common prefixes/suffixes
    name1 = this._normalizeName(name1);
    name2 = this._normalizeName(name2);

    const maxLength = Math.max(name1.length, name2.length);
    if (maxLength === 0) return 1;

    const distance = this.levenshteinDistance(name1, name2);
    return 1 - distance / maxLength;
  }

  /**
   * Calculate Levenshtein distance between strings.
   * Uses single-row DP (O(min(m,n)) space) with optional early exit.
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @param {number} [maxDistance=Infinity] - Bail out early if exceeded
   * @returns {number} - Distance (or maxDistance+1 if exceeded)
   */
  levenshteinDistance(str1, str2, maxDistance = Infinity) {
    // Ensure str1 is the shorter string for O(min) space
    if (str1.length > str2.length) {
      [str1, str2] = [str2, str1];
    }

    const m = str1.length;
    const n = str2.length;

    // Quick exits
    if (m === 0) return n;
    if (n - m > maxDistance) return maxDistance + 1;

    const prev = new Array(m + 1);
    const curr = new Array(m + 1);

    for (let i = 0; i <= m; i++) prev[i] = i;

    for (let j = 1; j <= n; j++) {
      curr[0] = j;
      let rowMin = j;

      for (let i = 1; i <= m; i++) {
        const cost = str1[i - 1] !== str2[j - 1] ? 1 : 0;
        curr[i] = Math.min(prev[i] + 1, curr[i - 1] + 1, prev[i - 1] + cost);
        if (curr[i] < rowMin) rowMin = curr[i];
      }

      // Early exit: if every value in this row exceeds the threshold,
      // the final distance will too
      if (rowMin > maxDistance) return maxDistance + 1;

      // Swap rows
      for (let i = 0; i <= m; i++) prev[i] = curr[i];
    }

    return prev[m];
  }

  /**
   * Compare two screenshots
   * @param {string} cypressShot - Cypress screenshot path
   * @param {string} playwrightShot - Playwright screenshot path
   */
  async compareScreenshots(cypressShot, playwrightShot) {
    try {
      await loadDeps();

      // Read images
      const cypress = _PNG.sync.read(await fs.readFile(cypressShot));
      const playwright = _PNG.sync.read(await fs.readFile(playwrightShot));

      // Check dimensions
      if (
        cypress.width !== playwright.width ||
        cypress.height !== playwright.height
      ) {
        // Resize images if necessary
        const { width, height } = this.calculateCommonDimensions(
          cypress,
          playwright
        );
        const resizedCypress = this.resizeImage(cypress, width, height);
        const resizedPlaywright = this.resizeImage(playwright, width, height);
        return this.performComparison(
          resizedCypress,
          resizedPlaywright,
          cypressShot,
          playwrightShot
        );
      }

      return this.performComparison(
        cypress,
        playwright,
        cypressShot,
        playwrightShot
      );
    } catch (error) {
      this.results.errors.push({
        type: 'comparison',
        cypressShot,
        playwrightShot,
        message: error.message,
      });
    }
  }

  /**
   * Calculate common dimensions for two images
   * @param {PNG} img1 - First image
   * @param {PNG} img2 - Second image
   * @returns {Object} - Common dimensions
   */
  calculateCommonDimensions(img1, img2) {
    return {
      width: Math.min(img1.width, img2.width),
      height: Math.min(img1.height, img2.height),
    };
  }

  /**
   * Resize PNG image
   * @param {PNG} image - Image to resize
   * @param {number} width - Target width
   * @param {number} height - Target height
   * @returns {PNG} - Resized image
   */
  resizeImage(image, width, height) {
    const resized = new _PNG({ width, height });
    // Simple nearest-neighbor scaling
    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        const srcX = Math.floor((x * image.width) / width);
        const srcY = Math.floor((y * image.height) / height);
        const srcIdx = (srcY * image.width + srcX) * 4;
        const destIdx = (y * width + x) * 4;
        resized.data[destIdx] = image.data[srcIdx];
        resized.data[destIdx + 1] = image.data[srcIdx + 1];
        resized.data[destIdx + 2] = image.data[srcIdx + 2];
        resized.data[destIdx + 3] = image.data[srcIdx + 3];
      }
    }
    return resized;
  }

  /**
   * Perform pixel-by-pixel comparison
   * @param {PNG} img1 - First image
   * @param {PNG} img2 - Second image
   * @param {string} cypressShotPath - Cypress screenshot path
   * @param {string} playwrightShotPath - Playwright screenshot path
   */
  async performComparison(img1, img2, cypressShotPath, playwrightShotPath) {
    const { width, height } = img1;
    const diff = new _PNG({ width, height });

    const mismatchedPixels = _pixelmatch(
      img1.data,
      img2.data,
      diff.data,
      width,
      height,
      {
        threshold: this.options.threshold,
        includeAA: true,
      }
    );

    const diffRatio = mismatchedPixels / (width * height);
    const diffPath = path.join(
      this.options.snapshotDir,
      `diff_${path.basename(cypressShotPath)}`
    );

    // Save diff image if there are differences
    if (diffRatio > 0 && this.options.saveSnapshots) {
      await fs.writeFile(diffPath, _PNG.sync.write(diff));
    }

    // Record result
    if (diffRatio <= this.options.threshold) {
      this.results.matches++;
    } else {
      this.results.mismatches++;
    }

    this.results.comparisons.push({
      cypressShot: cypressShotPath,
      playwrightShot: playwrightShotPath,
      diffRatio,
      diffPath: diffRatio > 0 ? diffPath : null,
      passed: diffRatio <= this.options.threshold,
    });
  }

  /**
   * Generate comparison report
   * @returns {Promise<Object>} - Comparison report
   */
  async generateReport() {
    const report = {
      summary: {
        total: this.results.matches + this.results.mismatches,
        matches: this.results.matches,
        mismatches: this.results.mismatches,
        errors: this.results.errors.length,
      },
      comparisons: this.results.comparisons,
      errors: this.results.errors,
      timestamp: new Date().toISOString(),
    };

    if (this.options.includeLogs) {
      // Generate HTML report
      const htmlReport = await this.generateHtmlReport(report);
      const reportPath = path.join(
        this.options.snapshotDir,
        'visual-report.html'
      );
      await fs.writeFile(reportPath, htmlReport);
    }

    return report;
  }

  /**
   * Generate HTML report
   * @param {Object} report - Comparison report
   * @returns {Promise<string>} - HTML content
   */
  async generateHtmlReport(report) {
    return `
<!DOCTYPE html>
<html>
<head>
  <title>Visual Comparison Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 2rem; }
    .summary { margin-bottom: 2rem; }
    .comparison { margin-bottom: 2rem; }
    .diff { display: flex; margin-top: 1rem; }
    .diff img { max-width: 300px; margin-right: 1rem; }
    .passed { color: green; }
    .failed { color: red; }
  </style>
</head>
<body>
  <h1>Visual Comparison Report</h1>
  
  <div class="summary">
    <h2>Summary</h2>
    <p>Total Comparisons: ${report.summary.total}</p>
    <p>Matches: ${report.summary.matches}</p>
    <p>Mismatches: ${report.summary.mismatches}</p>
    <p>Errors: ${report.summary.errors}</p>
  </div>

  <div class="comparisons">
    <h2>Detailed Results</h2>
    ${report.comparisons
      .map(
        (comp) => `
      <div class="comparison">
        <h3 class="${comp.passed ? 'passed' : 'failed'}">
          ${path.basename(comp.cypressShot)}
          (${(comp.diffRatio * 100).toFixed(2)}% difference)
        </h3>
        ${
          comp.diffPath
            ? `
          <div class="diff">
            <img src="${comp.cypressShot}" alt="Cypress version">
            <img src="${comp.playwrightShot}" alt="Playwright version">
            <img src="${comp.diffPath}" alt="Difference">
          </div>
        `
            : ''
        }
      </div>
    `
      )
      .join('')}
  </div>

  ${
    report.errors.length > 0
      ? `
    <div class="errors">
      <h2>Errors</h2>
      ${report.errors
        .map(
          (error) => `
        <div class="error">
          <p>Type: ${error.type}</p>
          <p>Message: ${error.message}</p>
        </div>
      `
        )
        .join('')}
    </div>
  `
      : ''
  }
</body>
</html>`;
  }

  /**
   * Log comparison summary
   */
  logSummary() {
    const { summary } = this.results;
    console.log('\nComparison Summary:');
    console.log(chalk.green(`Matches: ${summary.matches}`));
    console.log(chalk.red(`Mismatches: ${summary.mismatches}`));
    console.log(chalk.yellow(`Errors: ${summary.errors.length}`));
  }
}
