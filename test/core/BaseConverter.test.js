import { BaseConverter } from '../../src/core/BaseConverter.js';

describe('BaseConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new BaseConverter();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(converter.options.preserveComments).toBe(true);
      expect(converter.options.addTypeAnnotations).toBe(false);
    });

    it('should merge provided options with defaults', () => {
      const custom = new BaseConverter({ preserveComments: false, custom: 'value' });
      expect(custom.options.preserveComments).toBe(false);
      expect(custom.options.addTypeAnnotations).toBe(false);
      expect(custom.options.custom).toBe('value');
    });

    it('should initialize sourceFramework and targetFramework as null', () => {
      expect(converter.sourceFramework).toBeNull();
      expect(converter.targetFramework).toBeNull();
    });

    it('should initialize empty stats', () => {
      expect(converter.stats.conversions).toBe(0);
      expect(converter.stats.warnings).toEqual([]);
      expect(converter.stats.errors).toEqual([]);
    });
  });

  describe('abstract methods', () => {
    it('should throw when convert() is called', async () => {
      await expect(converter.convert('content')).rejects.toThrow(
        'convert() must be implemented by subclass'
      );
    });

    it('should throw when convertConfig() is called', async () => {
      await expect(converter.convertConfig('/path')).rejects.toThrow(
        'convertConfig() must be implemented by subclass'
      );
    });

    it('should throw when getImports() is called', () => {
      expect(() => converter.getImports()).toThrow(
        'getImports() must be implemented by subclass'
      );
    });

    it('should throw when detectTestTypes() is called', () => {
      expect(() => converter.detectTestTypes('content')).toThrow(
        'detectTestTypes() must be implemented by subclass'
      );
    });
  });

  describe('validate', () => {
    it('should return valid for correct JavaScript', () => {
      const result = converter.validate('const x = 1; const y = x + 2;');
      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
    });

    it('should return invalid for syntax errors', () => {
      const result = converter.validate('const x = {;');
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]).toContain('Syntax error');
    });
  });

  describe('getStats', () => {
    it('should return a copy of stats', () => {
      const stats = converter.getStats();
      stats.conversions = 999;
      expect(converter.stats.conversions).toBe(0);
    });
  });

  describe('reset', () => {
    it('should reset all stats to initial values', () => {
      converter.stats.conversions = 5;
      converter.stats.warnings.push({ message: 'test' });
      converter.stats.errors.push({ message: 'test' });

      converter.reset();

      expect(converter.stats.conversions).toBe(0);
      expect(converter.stats.warnings).toEqual([]);
      expect(converter.stats.errors).toEqual([]);
    });
  });

  describe('addWarning', () => {
    it('should add a warning with message and context', () => {
      converter.addWarning('test warning', { file: 'test.js' });
      expect(converter.stats.warnings).toHaveLength(1);
      expect(converter.stats.warnings[0].message).toBe('test warning');
      expect(converter.stats.warnings[0].context.file).toBe('test.js');
      expect(converter.stats.warnings[0].timestamp).toBeDefined();
    });

    it('should use empty context by default', () => {
      converter.addWarning('simple warning');
      expect(converter.stats.warnings[0].context).toEqual({});
    });
  });

  describe('addError', () => {
    it('should add an error with message and context', () => {
      converter.addError('test error', { line: 42 });
      expect(converter.stats.errors).toHaveLength(1);
      expect(converter.stats.errors[0].message).toBe('test error');
      expect(converter.stats.errors[0].context.line).toBe(42);
      expect(converter.stats.errors[0].timestamp).toBeDefined();
    });
  });

  describe('getSourceFramework', () => {
    it('should return the sourceFramework value', () => {
      converter.sourceFramework = 'cypress';
      expect(converter.getSourceFramework()).toBe('cypress');
    });
  });

  describe('getTargetFramework', () => {
    it('should return the targetFramework value', () => {
      converter.targetFramework = 'playwright';
      expect(converter.getTargetFramework()).toBe('playwright');
    });
  });

  describe('getConversionDirection', () => {
    it('should return formatted conversion direction', () => {
      converter.sourceFramework = 'cypress';
      converter.targetFramework = 'playwright';
      expect(converter.getConversionDirection()).toBe('cypress-to-playwright');
    });

    it('should handle null frameworks', () => {
      expect(converter.getConversionDirection()).toBe('null-to-null');
    });
  });
});
