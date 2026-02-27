import { VisualComparison } from '../../src/converter/visual.js';

describe('VisualComparison', () => {
  let comparison;

  beforeEach(() => {
    comparison = new VisualComparison();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(comparison.options.threshold).toBe(0.1);
      expect(comparison.options.includeLogs).toBe(true);
      expect(comparison.options.saveSnapshots).toBe(true);
      expect(comparison.options.snapshotDir).toBe('snapshots');
    });

    it('should accept custom options', () => {
      const custom = new VisualComparison({
        threshold: 0.05,
        includeLogs: false,
        saveSnapshots: false,
        snapshotDir: 'custom-snapshots',
      });
      expect(custom.options.threshold).toBe(0.05);
      expect(custom.options.includeLogs).toBe(false);
      expect(custom.options.saveSnapshots).toBe(false);
      expect(custom.options.snapshotDir).toBe('custom-snapshots');
    });

    it('should initialize empty results', () => {
      expect(comparison.results.comparisons).toEqual([]);
      expect(comparison.results.matches).toBe(0);
      expect(comparison.results.mismatches).toBe(0);
      expect(comparison.results.errors).toEqual([]);
    });
  });

  describe('lazy dependency loading', () => {
    it('should not require pngjs/pixelmatch for name-matching operations', () => {
      const vc = new VisualComparison();
      // These methods must work without triggering image dep loading
      const match = vc.findMatchingScreenshot('/shots/cypress-login.png', [
        '/shots/playwright-login.png',
      ]);
      expect(match).toBe('/shots/playwright-login.png');
      expect(vc.calculateNameSimilarity('login', 'login')).toBe(1);
      expect(vc.levenshteinDistance('abc', 'abc')).toBe(0);
    });
  });

  describe('calculateNameSimilarity', () => {
    it('should return 1 for identical names', () => {
      expect(comparison.calculateNameSimilarity('login', 'login')).toBe(1);
    });

    it('should strip cypress/playwright prefixes', () => {
      const similarity = comparison.calculateNameSimilarity(
        'cypress-login',
        'playwright-login'
      );
      expect(similarity).toBe(1);
    });

    it('should strip screenshot suffixes', () => {
      const similarity = comparison.calculateNameSimilarity(
        'login-screenshot',
        'login'
      );
      expect(similarity).toBe(1);
    });

    it('should return high similarity for similar names', () => {
      const similarity = comparison.calculateNameSimilarity(
        'login-page',
        'login-pag'
      );
      expect(similarity).toBeGreaterThan(0.8);
    });

    it('should return low similarity for different names', () => {
      const similarity = comparison.calculateNameSimilarity(
        'login',
        'dashboard'
      );
      expect(similarity).toBeLessThan(0.5);
    });
  });

  describe('levenshteinDistance', () => {
    it('should return 0 for identical strings', () => {
      expect(comparison.levenshteinDistance('hello', 'hello')).toBe(0);
    });

    it('should return correct distance for single edit', () => {
      expect(comparison.levenshteinDistance('hello', 'hallo')).toBe(1);
    });

    it('should return string length for completely different strings', () => {
      expect(comparison.levenshteinDistance('abc', 'xyz')).toBe(3);
    });

    it('should handle empty strings', () => {
      expect(comparison.levenshteinDistance('', 'abc')).toBe(3);
      expect(comparison.levenshteinDistance('abc', '')).toBe(3);
      expect(comparison.levenshteinDistance('', '')).toBe(0);
    });

    it('should handle insertions and deletions', () => {
      expect(comparison.levenshteinDistance('abc', 'abcd')).toBe(1);
      expect(comparison.levenshteinDistance('abcd', 'abc')).toBe(1);
    });

    it('should bail out early when maxDistance is exceeded', () => {
      // "abc" vs "xyz" has distance 3, maxDistance 1 should bail
      const result = comparison.levenshteinDistance('abc', 'xyz', 1);
      expect(result).toBeGreaterThan(1);
    });

    it('should return exact distance when within maxDistance', () => {
      expect(comparison.levenshteinDistance('hello', 'hallo', 5)).toBe(1);
    });

    it('should handle str2 shorter than str1 (swap)', () => {
      expect(comparison.levenshteinDistance('abcdef', 'abc')).toBe(3);
    });
  });

  describe('findMatchingScreenshot', () => {
    it('should find matching screenshot by name similarity', () => {
      const cypressShot = '/screenshots/cypress-login.png';
      const playwrightShots = [
        '/screenshots/playwright-login.png',
        '/screenshots/playwright-dashboard.png',
      ];
      const match = comparison.findMatchingScreenshot(
        cypressShot,
        playwrightShots
      );
      expect(match).toBe('/screenshots/playwright-login.png');
    });

    it('should return undefined when no match found', () => {
      const cypressShot = '/screenshots/cypress-login.png';
      const playwrightShots = ['/screenshots/playwright-dashboard.png'];
      const match = comparison.findMatchingScreenshot(
        cypressShot,
        playwrightShots
      );
      expect(match).toBeUndefined();
    });

    it('should handle empty playwright shots array', () => {
      const match = comparison.findMatchingScreenshot(
        '/screenshots/login.png',
        []
      );
      expect(match).toBeUndefined();
    });
  });

  describe('calculateCommonDimensions', () => {
    it('should return minimum dimensions', () => {
      const img1 = { width: 1920, height: 1080 };
      const img2 = { width: 1280, height: 720 };
      const result = comparison.calculateCommonDimensions(img1, img2);
      expect(result.width).toBe(1280);
      expect(result.height).toBe(720);
    });

    it('should return same dimensions when images are equal size', () => {
      const img1 = { width: 800, height: 600 };
      const img2 = { width: 800, height: 600 };
      const result = comparison.calculateCommonDimensions(img1, img2);
      expect(result.width).toBe(800);
      expect(result.height).toBe(600);
    });
  });

  describe('generateReport', () => {
    it('should return report with summary', async () => {
      comparison.options.includeLogs = false;
      comparison.results.matches = 5;
      comparison.results.mismatches = 2;
      comparison.results.errors.push({ type: 'missing', message: 'no match' });

      const report = await comparison.generateReport();
      expect(report.summary.total).toBe(7);
      expect(report.summary.matches).toBe(5);
      expect(report.summary.mismatches).toBe(2);
      expect(report.summary.errors).toBe(1);
      expect(report.timestamp).toBeDefined();
    });

    it('should include comparisons in report', async () => {
      comparison.options.includeLogs = false;
      comparison.results.comparisons.push({
        cypressShot: '/a.png',
        playwrightShot: '/b.png',
        diffRatio: 0.05,
        passed: true,
      });

      const report = await comparison.generateReport();
      expect(report.comparisons).toHaveLength(1);
      expect(report.comparisons[0].passed).toBe(true);
    });
  });

  describe('generateHtmlReport', () => {
    it('should generate valid HTML report', async () => {
      const report = {
        summary: { total: 2, matches: 1, mismatches: 1, errors: 0 },
        comparisons: [
          {
            cypressShot: '/cypress/login.png',
            playwrightShot: '/playwright/login.png',
            diffRatio: 0,
            diffPath: null,
            passed: true,
          },
        ],
        errors: [],
      };

      const html = await comparison.generateHtmlReport(report);
      expect(html).toContain('<!DOCTYPE html>');
      expect(html).toContain('Visual Comparison Report');
      expect(html).toContain('Total Comparisons: 2');
      expect(html).toContain('Matches: 1');
      expect(html).toContain('</html>');
    });

    it('should include error section when errors exist', async () => {
      const report = {
        summary: { total: 0, matches: 0, mismatches: 0, errors: 1 },
        comparisons: [],
        errors: [{ type: 'missing', message: 'No matching screenshot' }],
      };

      const html = await comparison.generateHtmlReport(report);
      expect(html).toContain('Errors');
      expect(html).toContain('No matching screenshot');
    });
  });
});
