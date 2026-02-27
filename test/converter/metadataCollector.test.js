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
      expect(collector.detectTestType('cy.mount(<Component />)')).toBe(
        'component'
      );
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
      expect(collector.getAssertionType('should("be.visible")')).toBe(
        'visibility'
      );
    });

    it('should identify text assertions', () => {
      expect(collector.getAssertionType('should("have.text", "hello")')).toBe(
        'text'
      );
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

  describe('extractCoverage', () => {
    it('should extract all coverage information', () => {
      const content = `
        cy.get('.btn-primary');
        cy.visit('/home');
        cy.click();
        expect(x).should('exist');
      `;
      const coverage = collector.extractCoverage(content);
      expect(coverage.selectors).toBeDefined();
      expect(coverage.routes).toBeDefined();
      expect(coverage.assertions).toBeDefined();
      expect(coverage.interactions).toBeDefined();
    });

    it('should return empty arrays for content without coverage items', () => {
      const coverage = collector.extractCoverage('const x = 1;');
      expect(coverage.selectors).toEqual([]);
      expect(coverage.routes).toEqual([]);
      expect(coverage.assertions).toEqual([]);
      expect(coverage.interactions).toEqual([]);
    });
  });

  describe('extractAssertions', () => {
    it('should extract assertion types and values', () => {
      const content = `should('exist'); should('be.visible'); should('have.text', 'hello');`;
      const assertions = collector.extractAssertions(content);
      expect(assertions.length).toBeGreaterThan(0);
      expect(assertions[0].type).toBeDefined();
      expect(assertions[0].value).toBeDefined();
    });

    it('should return empty array for no assertions', () => {
      expect(collector.extractAssertions('const x = 1;')).toEqual([]);
    });
  });

  describe('extractInteractions', () => {
    it('should extract click interactions', () => {
      const content = 'cy.click(); cy.type("hello");';
      const interactions = collector.extractInteractions(content);
      expect(interactions.length).toBeGreaterThan(0);
    });

    it('should extract multiple interaction types', () => {
      const content =
        'cy.click(); cy.type("hello"); cy.select("option"); cy.check(); cy.hover();';
      const interactions = collector.extractInteractions(content);
      expect(interactions.length).toBe(5);
    });

    it('should return empty array for no interactions', () => {
      expect(collector.extractInteractions('const x = 1;')).toEqual([]);
    });
  });

  describe('getAllTags', () => {
    it('should return empty array when no tags collected', () => {
      expect(collector.getAllTags()).toEqual([]);
    });
  });

  describe('summarizeTypes', () => {
    it('should count tests by type', () => {
      const tests = [
        { type: 'e2e' },
        { type: 'e2e' },
        { type: 'api' },
        { type: 'component' },
      ];
      const summary = collector.summarizeTypes(tests);
      expect(summary.e2e).toBe(2);
      expect(summary.api).toBe(1);
      expect(summary.component).toBe(1);
    });

    it('should return empty object for no tests', () => {
      expect(collector.summarizeTypes([])).toEqual({});
    });
  });

  describe('summarizeTags', () => {
    it('should count tags across tests', () => {
      const tests = [
        { tags: ['smoke', 'regression'] },
        { tags: ['smoke', 'e2e'] },
        { tags: ['regression'] },
      ];
      const summary = collector.summarizeTags(tests);
      expect(summary.smoke).toBe(2);
      expect(summary.regression).toBe(2);
      expect(summary.e2e).toBe(1);
    });

    it('should return empty object for no tests', () => {
      expect(collector.summarizeTags([])).toEqual({});
    });
  });

  describe('summarizeComplexity', () => {
    it('should calculate average complexity metrics', () => {
      const tests = [
        {
          complexity: { assertions: 2, commands: 4, conditionals: 0, hooks: 1 },
        },
        {
          complexity: { assertions: 4, commands: 6, conditionals: 2, hooks: 3 },
        },
      ];
      const summary = collector.summarizeComplexity(tests);
      expect(summary.averageAssertions).toBe(3);
      expect(summary.averageCommands).toBe(5);
      expect(summary.averageConditionals).toBe(1);
      expect(summary.averageHooks).toBe(2);
    });

    it('should handle empty tests array', () => {
      const summary = collector.summarizeComplexity([]);
      expect(summary.averageAssertions).toBe(0);
      expect(summary.averageCommands).toBe(0);
    });
  });

  describe('generateReport', () => {
    it('should generate report with tests from metadata', () => {
      collector.metadata.set('/test.js', {
        path: '/test.js',
        type: 'e2e',
        tags: ['smoke'],
        suites: [{ name: 'Suite' }],
        cases: [{ name: 'Test' }],
        complexity: { assertions: 2, commands: 3, conditionals: 0, hooks: 1 },
      });

      const report = collector.generateReport();
      expect(report.summary.totalTests).toBe(1);
      expect(report.summary.types.e2e).toBe(1);
      expect(report.tests).toHaveLength(1);
      expect(report.tests[0].path).toBe('/test.js');
    });
  });

  describe('collectMetadataFromContent', () => {
    // Use a real file path so getLastModified() can stat it
    const realPath = new URL(import.meta.url).pathname;

    it('should collect metadata from pre-read content', async () => {
      const content = `
        describe('Login', () => {
          it('should visit page', () => {
            cy.visit('/login');
            expect(true).toBe(true);
          });
        });
      `;
      const metadata = await collector.collectMetadataFromContent(
        realPath,
        content
      );
      expect(metadata.path).toBe(realPath);
      expect(metadata.type).toBe('e2e');
      expect(metadata.suites).toHaveLength(1);
      expect(metadata.cases).toHaveLength(1);
    });

    it('should produce same result as collectMetadata would for same content', async () => {
      const content = `
        describe('API test', () => {
          it('makes request', () => {
            cy.request('/api/data');
          });
        });
      `;
      const result = await collector.collectMetadataFromContent(
        realPath,
        content
      );
      expect(result.type).toBe('api');
      expect(result.suites[0].name).toBe('API test');
      expect(result.cases[0].name).toBe('makes request');
    });

    it('should store metadata in the internal map', async () => {
      const content = 'const x = 1;';
      await collector.collectMetadataFromContent(realPath, content);
      expect(collector.getMetadata(realPath)).not.toBeNull();
    });
  });
});
