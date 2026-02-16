import { DependencyAnalyzer } from '../../src/converter/dependencyAnalyzer.js';

describe('DependencyAnalyzer', () => {
  let analyzer;

  beforeEach(() => {
    analyzer = new DependencyAnalyzer();
  });

  describe('constructor', () => {
    it('should initialize with empty dependencies', () => {
      expect(analyzer.dependencies).toBeInstanceOf(Map);
      expect(analyzer.dependencies.size).toBe(0);
    });

    it('should initialize with empty import map', () => {
      expect(analyzer.importMap).toBeInstanceOf(Map);
    });

    it('should initialize with empty custom commands', () => {
      expect(analyzer.customCommands).toBeInstanceOf(Set);
      expect(analyzer.customCommands.size).toBe(0);
    });
  });

  describe('extractImports', () => {
    it('should extract named imports', () => {
      const content = "import { foo, bar } from './module.js';";
      const imports = analyzer.extractImports(content);
      expect(imports).toHaveLength(1);
      expect(imports[0].source).toBe('./module.js');
      expect(imports[0].specifiers).toContain('foo');
      expect(imports[0].specifiers).toContain('bar');
    });

    it('should extract default imports', () => {
      const content = "import fs from 'fs';";
      const imports = analyzer.extractImports(content);
      expect(imports).toHaveLength(1);
      expect(imports[0].source).toBe('fs');
    });

    it('should handle multiple imports', () => {
      const content = `
        import { test } from '@playwright/test';
        import path from 'path';
      `;
      const imports = analyzer.extractImports(content);
      expect(imports).toHaveLength(2);
    });

    it('should return empty for no imports', () => {
      expect(analyzer.extractImports('const x = 1;')).toEqual([]);
    });
  });

  describe('extractImportSpecifiers', () => {
    it('should extract specifiers from named import', () => {
      const specifiers = analyzer.extractImportSpecifiers("import { a, b, c } from 'mod'");
      expect(specifiers).toEqual(['a', 'b', 'c']);
    });

    it('should return empty for default import', () => {
      const specifiers = analyzer.extractImportSpecifiers("import x from 'mod'");
      expect(specifiers).toEqual([]);
    });
  });

  describe('extractCustomCommands', () => {
    it('should extract Cypress custom commands', () => {
      const content = `
        Cypress.Commands.add('login', { prevSubject: false }, function(email, password) {});
      `;
      const commands = analyzer.extractCustomCommands(content);
      expect(commands).toHaveLength(1);
      expect(commands[0].name).toBe('login');
    });

    it('should return empty when no custom commands', () => {
      expect(analyzer.extractCustomCommands('const x = 1;')).toEqual([]);
    });
  });

  describe('extractFixtures', () => {
    it('should extract cy.fixture calls', () => {
      const content = "cy.fixture('users.json'); cy.fixture('config.json');";
      const fixtures = analyzer.extractFixtures(content);
      expect(fixtures).toHaveLength(2);
      expect(fixtures[0].name).toBe('users.json');
      expect(fixtures[1].name).toBe('config.json');
    });

    it('should return empty when no fixtures', () => {
      expect(analyzer.extractFixtures('const x = 1;')).toEqual([]);
    });
  });

  describe('extractPageObjects', () => {
    it('should extract class definitions', () => {
      const content = `
        class LoginPage {
          constructor(page) {
            this.page = page;
          }
        }
      `;
      const pageObjects = analyzer.extractPageObjects(content);
      expect(pageObjects).toHaveLength(1);
      expect(pageObjects[0].name).toBe('LoginPage');
    });
  });

  describe('extractDependencies', () => {
    it('should extract require calls', () => {
      const content = "require('selenium-webdriver'); import('module');";
      const deps = analyzer.extractDependencies(content);
      expect(deps).toHaveLength(2);
      expect(deps[0].module).toBe('selenium-webdriver');
    });
  });

  describe('getDependencyTree', () => {
    it('should return null for unanalyzed path', () => {
      expect(analyzer.getDependencyTree('/nonexistent')).toBeNull();
    });

    it('should return dependency tree for analyzed path', () => {
      analyzer.dependencies.set('/test.js', {
        imports: [{ source: 'module', specifiers: [] }],
        customCommands: [],
        fixtures: [],
        pageObjects: [],
        dependencies: []
      });

      const tree = analyzer.getDependencyTree('/test.js');
      expect(tree).not.toBeNull();
      expect(tree.test).toBe('/test.js');
      expect(tree.imports).toHaveLength(1);
    });
  });

  describe('generateImportMap', () => {
    it('should return Cypress to Playwright import mappings', () => {
      const importMap = analyzer.generateImportMap();
      expect(importMap).toBeInstanceOf(Map);
      expect(importMap.get('@cypress/react')).toBe('@playwright/experimental-ct-react');
      expect(importMap.get('cypress-axe')).toBe('axe-playwright');
    });
  });

  describe('generateReport', () => {
    it('should return empty summary when no dependencies analyzed', () => {
      const report = analyzer.generateReport();
      expect(report.summary.totalFiles).toBe(0);
      expect(report.summary.totalImports).toBe(0);
      expect(report.summary.totalCustomCommands).toBe(0);
      expect(report.summary.totalFixtures).toBe(0);
      expect(report.files).toEqual([]);
    });

    it('should return summary of analyzed dependencies', () => {
      analyzer.dependencies.set('/test1.js', {
        imports: [{ source: 'module-a', specifiers: ['a'] }],
        customCommands: [{ name: 'login' }],
        fixtures: [{ name: 'users.json' }],
        pageObjects: [],
        dependencies: [],
      });
      analyzer.dependencies.set('/test2.js', {
        imports: [
          { source: 'module-b', specifiers: ['b'] },
          { source: 'module-c', specifiers: ['c'] },
        ],
        customCommands: [],
        fixtures: [],
        pageObjects: [{ name: 'LoginPage' }],
        dependencies: [{ module: 'lodash' }],
      });

      const report = analyzer.generateReport();
      expect(report.summary.totalFiles).toBe(2);
      expect(report.summary.totalImports).toBe(3);
      expect(report.summary.totalCustomCommands).toBe(1);
      expect(report.summary.totalFixtures).toBe(1);
      expect(report.files).toHaveLength(2);
      expect(report.files[0].path).toBe('/test1.js');
      expect(report.files[1].path).toBe('/test2.js');
    });
  });

  describe('findCircularDependencies', () => {
    it('should return empty array when no circular deps', () => {
      analyzer.dependencies.set('a.js', { dependencies: [{ module: 'b.js' }] });
      analyzer.dependencies.set('b.js', { dependencies: [] });

      const circular = analyzer.findCircularDependencies('a.js');
      expect(circular).toEqual([]);
    });

    it('should detect circular dependencies', () => {
      analyzer.dependencies.set('a.js', { dependencies: [{ module: 'b.js' }] });
      analyzer.dependencies.set('b.js', { dependencies: [{ module: 'a.js' }] });

      const circular = analyzer.findCircularDependencies('a.js');
      expect(circular.length).toBeGreaterThan(0);
    });
  });
});
