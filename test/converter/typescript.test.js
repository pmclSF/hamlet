import { TypeScriptConverter } from '../../src/converter/typescript.js';

describe('TypeScriptConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new TypeScriptConverter();
  });

  describe('constructor', () => {
    it('should initialize with type mappings', () => {
      expect(converter.typeMap).toBeInstanceOf(Map);
      expect(converter.typeMap.size).toBeGreaterThan(0);
    });

    it('should initialize with interface mappings', () => {
      expect(converter.interfaceMap).toBeInstanceOf(Map);
      expect(converter.interfaceMap.size).toBeGreaterThan(0);
    });
  });

  describe('convertContent', () => {
    it('should replace Cypress.Chainable with Locator', () => {
      const input = 'const el: Cypress.Chainable = cy.get("#btn");';
      const result = converter.convertContent(input);
      expect(result).toContain('Locator');
      expect(result).not.toContain('Cypress.Chainable');
    });

    it('should replace Cypress.Config with PlaywrightTestConfig', () => {
      const input = 'const config: Cypress.Config = {};';
      const result = converter.convertContent(input);
      expect(result).toContain('PlaywrightTestConfig');
      expect(result).not.toContain('Cypress.Config');
    });

    it('should replace Cypress.Response with APIResponse', () => {
      const input = 'function handle(resp: Cypress.Response) {}';
      const result = converter.convertContent(input);
      expect(result).toContain('APIResponse');
      expect(result).not.toContain('Cypress.Response');
    });

    it('should replace interface names from interfaceMap', () => {
      const input = 'interface CypressConfiguration { baseUrl: string; }';
      const result = converter.convertContent(input);
      expect(result).toContain('PlaywrightTestConfig');
      expect(result).not.toContain('CypressConfiguration');
    });

    it('should handle content with no Cypress types', () => {
      const input = 'const x = 42;\nfunction hello() { return "world"; }';
      const result = converter.convertContent(input);
      expect(result).toBe(input);
    });

    it('should handle empty content', () => {
      const result = converter.convertContent('');
      expect(result).toBe('');
    });

    it('should replace multiple occurrences', () => {
      const input = 'Cypress.Chainable and Cypress.Chainable again';
      const result = converter.convertContent(input);
      expect(result).toBe('Locator and Locator again');
    });
  });

  describe('transformMethodName', () => {
    it('should map visit to goto', () => {
      expect(converter.transformMethodName('visit')).toBe('goto');
    });

    it('should map get to locator', () => {
      expect(converter.transformMethodName('get')).toBe('locator');
    });

    it('should map type to fill', () => {
      expect(converter.transformMethodName('type')).toBe('fill');
    });

    it('should return original name for unmapped methods', () => {
      expect(converter.transformMethodName('customMethod')).toBe('customMethod');
    });
  });
});
