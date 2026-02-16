import { TestMetadataCollector } from '../../src/converter/metadataCollector.js';

describe('TestMetadataCollector', () => {
  let collector;

  beforeEach(() => {
    collector = new TestMetadataCollector();
  });

  describe('constructor', () => {
    it('should initialize with empty metadata', () => {
      expect(collector.metadata).toBeInstanceOf(Map);
      expect(collector.metadata.size).toBe(0);
    });

    it('should initialize with empty test suites', () => {
      expect(collector.testSuites).toBeInstanceOf(Map);
    });

    it('should initialize with empty tags', () => {
      expect(collector.tags).toBeInstanceOf(Set);
      expect(collector.tags.size).toBe(0);
    });
  });

  describe('detectTestType', () => {
    it('should detect e2e tests', () => {
      expect(collector.detectTestType('cy.visit("/home")')).toBe('e2e');
    });

    it('should detect component tests', () => {
      expect(collector.detectTestType('cy.mount(<Component />)')).toBe('component');
    });

    it('should detect api tests', () => {
      expect(collector.detectTestType('cy.request("/api/users")')).toBe('api');
    });

    it('should detect visual tests', () => {
      expect(collector.detectTestType('matchImageSnapshot()')).toBe('visual');
    });

    it('should detect accessibility tests', () => {
      expect(collector.detectTestType('cy.checkA11y()')).toBe('accessibility');
    });

    it('should return unknown for unrecognized content', () => {
      expect(collector.detectTestType('const x = 1;')).toBe('unknown');
    });
  });

  describe('extractTestSuites', () => {
    it('should extract describe blocks', () => {
      const content = `
        describe('Login Suite', () => {
          describe('Nested Suite', () => {});
        });
      `;
      const suites = collector.extractTestSuites(content);
      expect(suites).toHaveLength(2);
      expect(suites[0].name).toBe('Login Suite');
      expect(suites[1].name).toBe('Nested Suite');
    });

    it('should return empty array for no suites', () => {
      expect(collector.extractTestSuites('const x = 1;')).toEqual([]);
    });
  });

  describe('extractTestCases', () => {
    it('should extract it blocks', () => {
      const content = `
        it('should login', () => {});
        it('should logout', () => {});
      `;
      const cases = collector.extractTestCases(content);
      expect(cases).toHaveLength(2);
      expect(cases[0].name).toBe('should login');
      expect(cases[1].name).toBe('should logout');
    });
  });

  describe('extractTags', () => {
    it('should extract known tags', () => {
      const content = '// @smoke @regression @e2e';
      const tags = collector.extractTags(content);
      expect(tags).toContain('smoke');
      expect(tags).toContain('regression');
      expect(tags).toContain('e2e');
    });

    it('should return empty for no tags', () => {
      expect(collector.extractTags('const x = 1;')).toEqual([]);
    });

    it('should deduplicate tags', () => {
      const content = '// @smoke @smoke @smoke';
      const tags = collector.extractTags(content);
      expect(tags).toHaveLength(1);
    });
  });

  describe('calculateComplexity', () => {
    it('should count assertions', () => {
      const content = 'expect(x).toBe(1); expect(y).toBe(2); should("exist");';
      const complexity = collector.calculateComplexity(content);
      expect(complexity.assertions).toBe(3);
    });

    it('should count cy commands', () => {
      const content = 'cy.visit("/"); cy.get(".btn"); cy.click();';
      const complexity = collector.calculateComplexity(content);
      expect(complexity.commands).toBe(3);
    });

    it('should count conditionals', () => {
      const content = 'if (x) { } else { } switch(y) { case 1: }';
      const complexity = collector.calculateComplexity(content);
      expect(complexity.conditionals).toBe(4);
    });

    it('should return zeros for simple content', () => {
      const complexity = collector.calculateComplexity('const x = 1;');
      expect(complexity.assertions).toBe(0);
      expect(complexity.commands).toBe(0);
    });
  });

  describe('extractSelectors', () => {
    it('should extract cy.get selectors', () => {
      const content = "cy.get('.btn-primary'); cy.get('#login');";
      const selectors = collector.extractSelectors(content);
      expect(selectors).toContain('.btn-primary');
      expect(selectors).toContain('#login');
    });

    it('should extract cy.contains selectors', () => {
      const content = "cy.contains('Submit');";
      const selectors = collector.extractSelectors(content);
      expect(selectors).toContain('Submit');
    });
  });

  describe('extractRoutes', () => {
    it('should extract cy.visit routes', () => {
      const content = "cy.visit('/home'); cy.visit('/about');";
      const routes = collector.extractRoutes(content);
      expect(routes).toContain('/home');
      expect(routes).toContain('/about');
    });

    it('should extract cy.intercept routes', () => {
      const content = "cy.intercept('/api/users');";
      const routes = collector.extractRoutes(content);
      expect(routes).toContain('/api/users');
    });
  });

  describe('getAssertionType', () => {
    it('should identify existence assertions', () => {
      expect(collector.getAssertionType('should("exist")')).toBe('existence');
    });

    it('should identify visibility assertions', () => {
      expect(collector.getAssertionType('should("be.visible")')).toBe('visibility');
    });

    it('should identify text assertions', () => {
      expect(collector.getAssertionType('should("have.text", "hello")')).toBe('text');
    });

    it('should return other for unrecognized assertions', () => {
      expect(collector.getAssertionType('should("be.checked")')).toBe('other');
    });
  });

  describe('getMetadata', () => {
    it('should return null for uncollected path', () => {
      expect(collector.getMetadata('/nonexistent')).toBeNull();
    });
  });

  describe('getTestsByTag', () => {
    it('should return empty array when no metadata matches', () => {
      expect(collector.getTestsByTag('smoke')).toEqual([]);
    });
  });

  describe('generateReport', () => {
    it('should return summary with zero counts when empty', () => {
      const report = collector.generateReport();
      expect(report.summary.totalTests).toBe(0);
      expect(report.tests).toEqual([]);
    });
  });

  describe('average', () => {
    it('should calculate average of numbers', () => {
      expect(collector.average([2, 4, 6])).toBe(4);
    });

    it('should return 0 for empty array', () => {
      expect(collector.average([])).toBe(0);
    });
  });
});
