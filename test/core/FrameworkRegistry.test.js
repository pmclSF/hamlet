import { FrameworkRegistry } from '../../src/core/FrameworkRegistry.js';

function makeDefinition(overrides = {}) {
  return {
    name: 'testfw',
    language: 'javascript',
    detect: () => 100,
    parse: () => ({}),
    emit: () => '',
    imports: {},
    paradigm: 'bdd',
    ...overrides,
  };
}

describe('FrameworkRegistry', () => {
  let registry;

  beforeEach(() => {
    registry = new FrameworkRegistry();
  });

  describe('register', () => {
    it('should register a valid framework definition', () => {
      const def = makeDefinition({ name: 'jest' });
      registry.register(def);
      expect(registry.has('jest')).toBe(true);
    });

    it('should reject definition missing name', () => {
      const def = makeDefinition();
      delete def.name;
      expect(() => registry.register(def)).toThrow(/missing required fields.*name/);
    });

    it('should reject definition missing language', () => {
      const def = makeDefinition();
      delete def.language;
      expect(() => registry.register(def)).toThrow(/missing required fields.*language/);
    });

    it('should reject definition missing detect', () => {
      const def = makeDefinition();
      delete def.detect;
      expect(() => registry.register(def)).toThrow(/missing required fields.*detect/);
    });

    it('should reject definition missing parse', () => {
      const def = makeDefinition();
      delete def.parse;
      expect(() => registry.register(def)).toThrow(/missing required fields.*parse/);
    });

    it('should reject definition missing emit', () => {
      const def = makeDefinition();
      delete def.emit;
      expect(() => registry.register(def)).toThrow(/missing required fields.*emit/);
    });

    it('should reject definition missing imports', () => {
      const def = makeDefinition();
      delete def.imports;
      expect(() => registry.register(def)).toThrow(/missing required fields.*imports/);
    });

    it('should reject definition missing paradigm', () => {
      const def = makeDefinition();
      delete def.paradigm;
      expect(() => registry.register(def)).toThrow(/missing required fields.*paradigm/);
    });

    it('should reject definition missing multiple fields', () => {
      const def = makeDefinition();
      delete def.name;
      delete def.language;
      expect(() => registry.register(def)).toThrow(/missing required fields.*name.*language/);
    });

    it('should overwrite existing registration for same (language, name)', () => {
      const def1 = makeDefinition({ name: 'jest', paradigm: 'bdd' });
      const def2 = makeDefinition({ name: 'jest', paradigm: 'xunit' });
      registry.register(def1);
      registry.register(def2);
      expect(registry.get('jest').paradigm).toBe('xunit');
    });
  });

  describe('get', () => {
    it('should retrieve a framework by name', () => {
      const def = makeDefinition({ name: 'jest' });
      registry.register(def);
      expect(registry.get('jest')).toBe(def);
    });

    it('should return null for unregistered framework', () => {
      expect(registry.get('nonexistent')).toBeNull();
    });

    it('should retrieve by (name, language)', () => {
      const def = makeDefinition({ name: 'jest', language: 'javascript' });
      registry.register(def);
      expect(registry.get('jest', 'javascript')).toBe(def);
    });

    it('should return null when language does not match', () => {
      const def = makeDefinition({ name: 'jest', language: 'javascript' });
      registry.register(def);
      expect(registry.get('jest', 'python')).toBeNull();
    });

    it('should disambiguate same name across languages', () => {
      const jsDef = makeDefinition({ name: 'selenium', language: 'javascript' });
      const pyDef = makeDefinition({ name: 'selenium', language: 'python' });
      registry.register(jsDef);
      registry.register(pyDef);

      expect(registry.get('selenium', 'javascript')).toBe(jsDef);
      expect(registry.get('selenium', 'python')).toBe(pyDef);
    });

    it('should return first match when no language specified with duplicates', () => {
      const jsDef = makeDefinition({ name: 'selenium', language: 'javascript' });
      const pyDef = makeDefinition({ name: 'selenium', language: 'python' });
      registry.register(jsDef);
      registry.register(pyDef);

      const result = registry.get('selenium');
      expect([jsDef, pyDef]).toContain(result);
    });
  });

  describe('has', () => {
    it('should return true for registered framework', () => {
      registry.register(makeDefinition({ name: 'jest' }));
      expect(registry.has('jest')).toBe(true);
    });

    it('should return false for unregistered framework', () => {
      expect(registry.has('jest')).toBe(false);
    });

    it('should scope by language', () => {
      registry.register(makeDefinition({ name: 'jest', language: 'javascript' }));
      expect(registry.has('jest', 'javascript')).toBe(true);
      expect(registry.has('jest', 'python')).toBe(false);
    });
  });

  describe('list', () => {
    it('should return empty array when no frameworks registered', () => {
      expect(registry.list()).toEqual([]);
    });

    it('should return all registered frameworks', () => {
      registry.register(makeDefinition({ name: 'jest', language: 'javascript' }));
      registry.register(makeDefinition({ name: 'vitest', language: 'javascript' }));
      registry.register(makeDefinition({ name: 'pytest', language: 'python' }));

      expect(registry.list()).toHaveLength(3);
    });

    it('should filter by language', () => {
      registry.register(makeDefinition({ name: 'jest', language: 'javascript' }));
      registry.register(makeDefinition({ name: 'vitest', language: 'javascript' }));
      registry.register(makeDefinition({ name: 'pytest', language: 'python' }));

      const jsFrameworks = registry.list('javascript');
      expect(jsFrameworks).toHaveLength(2);
      expect(jsFrameworks.map(f => f.name)).toEqual(
        expect.arrayContaining(['jest', 'vitest'])
      );

      const pyFrameworks = registry.list('python');
      expect(pyFrameworks).toHaveLength(1);
      expect(pyFrameworks[0].name).toBe('pytest');
    });

    it('should return empty array for language with no frameworks', () => {
      registry.register(makeDefinition({ name: 'jest', language: 'javascript' }));
      expect(registry.list('ruby')).toEqual([]);
    });
  });

  describe('clear', () => {
    it('should remove all registered frameworks', () => {
      registry.register(makeDefinition({ name: 'jest' }));
      registry.register(makeDefinition({ name: 'vitest' }));
      expect(registry.list()).toHaveLength(2);

      registry.clear();
      expect(registry.list()).toEqual([]);
      expect(registry.has('jest')).toBe(false);
    });
  });
});
