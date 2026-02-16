import { TestMapper } from '../../src/converter/mapper.js';

describe('TestMapper', () => {
  let mapper;

  beforeEach(() => {
    mapper = new TestMapper();
  });

  describe('constructor', () => {
    it('should initialize with empty mappings', () => {
      expect(mapper.mappings).toBeInstanceOf(Map);
      expect(mapper.mappings.size).toBe(0);
    });

    it('should initialize metadata', () => {
      expect(mapper.metaData.version).toBe('1.0');
      expect(mapper.metaData.statistics.totalMappings).toBe(0);
    });
  });

  describe('calculateNameSimilarity', () => {
    it('should return 1 for identical names', () => {
      expect(mapper.calculateNameSimilarity('login', 'login')).toBe(1);
    });

    it('should return high similarity for same base with different suffixes', () => {
      const similarity = mapper.calculateNameSimilarity('login.cy', 'login.spec');
      expect(similarity).toBeGreaterThan(0.5);
    });

    it('should strip test suffixes before comparing', () => {
      const similarity = mapper.calculateNameSimilarity('login.test', 'login.spec');
      expect(similarity).toBe(1);
    });
  });

  describe('levenshteinDistance', () => {
    it('should return 0 for identical strings', () => {
      expect(mapper.levenshteinDistance('hello', 'hello')).toBe(0);
    });

    it('should return correct distance for single edit', () => {
      expect(mapper.levenshteinDistance('hello', 'hallo')).toBe(1);
    });

    it('should return string length for completely different strings', () => {
      expect(mapper.levenshteinDistance('abc', 'xyz')).toBe(3);
    });

    it('should handle empty strings', () => {
      expect(mapper.levenshteinDistance('', 'abc')).toBe(3);
      expect(mapper.levenshteinDistance('abc', '')).toBe(3);
    });
  });

  describe('calculateContentSimilarity', () => {
    it('should return 1 for identical test descriptions', () => {
      const content = "describe('Login', () => { it('should login', () => {}) })";
      const similarity = mapper.calculateContentSimilarity(content, content);
      expect(similarity).toBe(1);
    });

    it('should return 0 for completely different tests', () => {
      const content1 = "describe('Login', () => {})";
      const content2 = "test('something completely different', () => {})";
      const similarity = mapper.calculateContentSimilarity(content1, content2);
      expect(similarity).toBeLessThan(1);
    });
  });

  describe('updateStatistics', () => {
    it('should update metadata statistics based on mappings', () => {
      mapper.mappings.set('test1.cy.js', {
        playwrightPath: 'test1.spec.js',
        status: 'active',
        syncStatus: 'synced'
      });
      mapper.mappings.set('test2.cy.js', {
        playwrightPath: 'test2.spec.js',
        status: 'active',
        syncStatus: 'pending'
      });

      mapper.updateStatistics();

      expect(mapper.metaData.statistics.totalMappings).toBe(2);
      expect(mapper.metaData.statistics.activeMappings).toBe(2);
      expect(mapper.metaData.statistics.pendingSync).toBe(1);
    });
  });

  describe('getMappings', () => {
    it('should return metadata and mapping arrays', () => {
      mapper.mappings.set('test.cy.js', {
        playwrightPath: 'test.spec.js',
        status: 'active',
        syncStatus: 'synced',
        lastSync: '2024-01-01'
      });

      const result = mapper.getMappings();
      expect(result.mappings).toHaveLength(1);
      expect(result.mappings[0].cypressTest).toBe('test.cy.js');
      expect(result.mappings[0].playwrightTest).toBe('test.spec.js');
    });

    it('should return empty mappings array when no mappings exist', () => {
      const result = mapper.getMappings();
      expect(result.mappings).toEqual([]);
    });
  });
});
