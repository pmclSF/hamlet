import fs from 'fs/promises';
import { logUtils } from '../utils/helpers.js';

const logger = logUtils.createLogger('DependencyAnalyzer');

/**
 * Analyzes and manages test dependencies
 */
export class DependencyAnalyzer {
  constructor() {
    this.dependencies = new Map();
    this.importMap = new Map();
    this.customCommands = new Set();
  }

  /**
   * Analyze test dependencies
   * @param {string} testPath - Path to test file
   * @returns {Promise<Object>} - Dependency analysis
   */
  async analyzeDependencies(testPath) {
    try {
      const content = await fs.readFile(testPath, 'utf8');
      
      const analysis = {
        imports: this.extractImports(content),
        customCommands: this.extractCustomCommands(content),
        fixtures: this.extractFixtures(content),
        pageObjects: this.extractPageObjects(content),
        dependencies: this.extractDependencies(content)
      };

      this.dependencies.set(testPath, analysis);
      return analysis;

    } catch (error) {
      logger.error(`Failed to analyze ${testPath}:`, error);
      throw error;
    }
  }

  /**
   * Extract import statements
   * @param {string} content - File content
   * @returns {Object[]} - Array of imports
   */
  extractImports(content) {
    const imports = [];
    const importRegex = /import\s+(?:{[^}]+}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]/g;
    
    let match;
    while ((match = importRegex.exec(content)) !== null) {
      imports.push({
        statement: match[0],
        source: match[1],
        specifiers: this.extractImportSpecifiers(match[0])
      });
    }

    return imports;
  }

  /**
   * Extract import specifiers
   * @param {string} importStatement - Import statement
   * @returns {string[]} - Array of imported specifiers
   */
  extractImportSpecifiers(importStatement) {
    const specifierRegex = /{([^}]+)}/;
    const match = importStatement.match(specifierRegex);
    
    if (match) {
      return match[1].split(',').map(s => s.trim());
    }
    
    return [];
  }

  /**
   * Extract custom commands
   * @param {string} content - File content
   * @returns {Object[]} - Array of custom commands
   */
  extractCustomCommands(content) {
    const commands = [];
    const commandRegex = /Cypress\.Commands\.add\(['"](.*?)['"],\s*(?:{\s*prevSubject:\s*(.*?)\s*})?,\s*function/g;
    
    let match;
    while ((match = commandRegex.exec(content)) !== null) {
      commands.push({
        name: match[1],
        chainable: match[2] === 'true',
        location: match.index
      });
    }

    return commands;
  }

  /**
   * Extract fixtures
   * @param {string} content - File content
   * @returns {Object[]} - Array of fixtures
   */
  extractFixtures(content) {
    const fixtures = [];
    const fixtureRegex = /cy\.fixture\(['"](.*?)['"]\)/g;
    
    let match;
    while ((match = fixtureRegex.exec(content)) !== null) {
      fixtures.push({
        name: match[1],
        location: match.index
      });
    }

    return fixtures;
  }

  /**
   * Extract page objects
   * @param {string} content - File content
   * @returns {Object[]} - Array of page objects
   */
  extractPageObjects(content) {
    const pageObjects = [];
    const pageObjectRegex = /class\s+(\w+)\s*{[\s\S]*?constructor\s*\([^)]*\)/g;
    
    let match;
    while ((match = pageObjectRegex.exec(content)) !== null) {
      pageObjects.push({
        name: match[1],
        location: match.index
      });
    }

    return pageObjects;
  }

  /**
   * Extract dependencies
   * @param {string} content - File content
   * @returns {Object[]} - Array of dependencies
   */
  extractDependencies(content) {
    const dependencies = [];
    const requireRegex = /(?:require|import)\s*\(['"](.*?)['"]\)/g;
    
    let match;
    while ((match = requireRegex.exec(content)) !== null) {
      dependencies.push({
        module: match[1],
        location: match.index
      });
    }

    return dependencies;
  }

  /**
   * Generate a summary report of all analyzed dependencies
   * @returns {Object} - Dependency report with summary and per-file details
   */
  generateReport() {
    const files = Array.from(this.dependencies.entries());
    const allImports = files.flatMap(([, a]) => a.imports || []);
    const allCustomCommands = files.flatMap(([, a]) => a.customCommands || []);
    const allFixtures = files.flatMap(([, a]) => a.fixtures || []);

    return {
      summary: {
        totalFiles: files.length,
        totalImports: allImports.length,
        totalCustomCommands: allCustomCommands.length,
        totalFixtures: allFixtures.length,
      },
      files: files.map(([filePath, analysis]) => ({
        path: filePath,
        imports: analysis.imports,
        customCommands: analysis.customCommands,
        fixtures: analysis.fixtures,
        pageObjects: analysis.pageObjects,
        dependencies: analysis.dependencies,
      })),
    };
  }

  /**
   * Get dependency tree for a test
   * @param {string} testPath - Path to test file
   * @returns {Object} - Dependency tree
   */
  getDependencyTree(testPath) {
    const analysis = this.dependencies.get(testPath);
    if (!analysis) {
      return null;
    }

    return {
      test: testPath,
      imports: analysis.imports,
      customCommands: analysis.customCommands,
      fixtures: analysis.fixtures,
      pageObjects: analysis.pageObjects,
      dependencies: analysis.dependencies
    };
  }

  /**
   * Generate import map for conversion
   * @returns {Map<string, string>} - Import mappings
   */
  generateImportMap() {
    const importMap = new Map();

    // Cypress to Playwright mappings
    importMap.set('@cypress/react', '@playwright/experimental-ct-react');
    importMap.set('cypress-axe', 'axe-playwright');
    importMap.set('cypress-file-upload', '@playwright/test');
    importMap.set('cypress-real-events', '@playwright/test');

    return importMap;
  }

  /**
   * Check for circular dependencies
   * @param {string} testPath - Path to test file
   * @returns {string[]} - Array of circular dependencies
   */
  findCircularDependencies(testPath) {
    const visited = new Set();
    const path = [];
    const circular = [];

    const visit = (file) => {
      if (path.includes(file)) {
        circular.push([...path.slice(path.indexOf(file)), file]);
        return;
      }
      
      if (visited.has(file)) return;
      
      visited.add(file);
      path.push(file);
      
      const analysis = this.dependencies.get(file);
      if (analysis) {
        for (const dep of analysis.dependencies) {
          visit(dep.module);
        }
      }
      
      path.pop();
    };

    visit(testPath);
    return circular;
  }
}