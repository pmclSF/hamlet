import { TypeScriptConverter } from '../../src/converter/typescript.js';
import ts from 'typescript';

describe('TypeScriptConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new TypeScriptConverter();
  });

  describe('constructor', () => {
    it('should initialize typeMap with Cypress to Playwright type mappings', () => {
      expect(converter.typeMap).toBeInstanceOf(Map);
      expect(converter.typeMap.get('Cypress.Chainable')).toBe('Locator');
      expect(converter.typeMap.get('cy.wrap')).toBe('Promise');
      expect(converter.typeMap.get('Cypress.Config')).toBe(
        'PlaywrightTestConfig'
      );
      expect(converter.typeMap.get('cy.stub')).toBe('Mock');
      expect(converter.typeMap.get('cy.spy')).toBe('Mock');
      expect(converter.typeMap.get('Cypress.Browser')).toBe('BrowserContext');
      expect(converter.typeMap.get('Cypress.ElementHandle')).toBe(
        'ElementHandle'
      );
      expect(converter.typeMap.get('Cypress.Cookie')).toBe('Cookie');
      expect(converter.typeMap.get('Cypress.Response')).toBe('APIResponse');
      expect(converter.typeMap.get('Cypress.AUTWindow')).toBe('Page');
    });

    it('should initialize interfaceMap with Cypress to Playwright interface mappings', () => {
      expect(converter.interfaceMap).toBeInstanceOf(Map);
      expect(converter.interfaceMap.get('CypressConfiguration')).toBe(
        'PlaywrightTestConfig'
      );
      expect(converter.interfaceMap.get('CypressPlugin')).toBe(
        'PlaywrightPlugin'
      );
      expect(converter.interfaceMap.get('CypressCommand')).toBe(
        'PlaywrightTest'
      );
      expect(converter.interfaceMap.get('CypressFixture')).toBe('TestFixture');
    });

    it('should have all 16 type mappings', () => {
      expect(converter.typeMap.size).toBe(16);
    });

    it('should have all 4 interface mappings', () => {
      expect(converter.interfaceMap.size).toBe(4);
    });
  });

  describe('transformMethodName', () => {
    it('should map visit to goto', () => {
      expect(converter.transformMethodName('visit')).toBe('goto');
    });

    it('should map get to locator', () => {
      expect(converter.transformMethodName('get')).toBe('locator');
    });

    it('should map find to locator', () => {
      expect(converter.transformMethodName('find')).toBe('locator');
    });

    it('should map type to fill', () => {
      expect(converter.transformMethodName('type')).toBe('fill');
    });

    it('should map should to expect', () => {
      expect(converter.transformMethodName('should')).toBe('expect');
    });

    it('should map contains to getByText', () => {
      expect(converter.transformMethodName('contains')).toBe('getByText');
    });

    it('should map first/last/eq correctly', () => {
      expect(converter.transformMethodName('first')).toBe('first');
      expect(converter.transformMethodName('last')).toBe('last');
      expect(converter.transformMethodName('eq')).toBe('nth');
    });

    it('should map parent to locator(..)', () => {
      expect(converter.transformMethodName('parent')).toBe("locator('..')");
    });

    it('should return original name for unmapped methods', () => {
      expect(converter.transformMethodName('customMethod')).toBe(
        'customMethod'
      );
      expect(converter.transformMethodName('doSomething')).toBe('doSomething');
    });

    it('should map click to click', () => {
      expect(converter.transformMethodName('click')).toBe('click');
    });

    it('should map siblings to locator(~)', () => {
      expect(converter.transformMethodName('siblings')).toBe("locator('~')");
    });

    it('should map next and prev', () => {
      expect(converter.transformMethodName('next')).toBe("locator('+')");
      expect(converter.transformMethodName('prev')).toBe("locator('-')");
    });

    it('should map children to locator(>*)', () => {
      expect(converter.transformMethodName('children')).toBe("locator('>*')");
    });
  });

  describe('generateDefinitionFileContent', () => {
    it('should generate header with Playwright imports', async () => {
      const content = await converter.generateDefinitionFileContent(new Map());
      expect(content).toContain(
        '// Generated type definitions for Playwright tests'
      );
      expect(content).toContain(
        "import { test, expect, Page, Locator } from '@playwright/test';"
      );
    });

    it('should generate interface declarations', async () => {
      const typeDefs = new Map();
      typeDefs.set('LoginPage', {
        kind: ts.SyntaxKind.InterfaceDeclaration,
        type: '{ username: string; password: string; }',
      });

      const content = await converter.generateDefinitionFileContent(typeDefs);
      expect(content).toContain('interface LoginPage');
    });

    it('should generate type alias declarations', async () => {
      const typeDefs = new Map();
      typeDefs.set('UserId', {
        kind: ts.SyntaxKind.TypeAliasDeclaration,
        type: 'string',
      });

      const content = await converter.generateDefinitionFileContent(typeDefs);
      expect(content).toContain('type UserId = string;');
    });

    it('should handle empty type definitions', async () => {
      const content = await converter.generateDefinitionFileContent(new Map());
      expect(content).toContain('// Generated type definitions');
    });

    it('should handle mixed definitions', async () => {
      const typeDefs = new Map();
      typeDefs.set('Config', {
        kind: ts.SyntaxKind.InterfaceDeclaration,
        type: '{ timeout: number; }',
      });
      typeDefs.set('TestId', {
        kind: ts.SyntaxKind.TypeAliasDeclaration,
        type: 'string | number',
      });

      const content = await converter.generateDefinitionFileContent(typeDefs);
      expect(content).toContain('interface Config');
      expect(content).toContain('type TestId = string | number;');
    });
  });
});
