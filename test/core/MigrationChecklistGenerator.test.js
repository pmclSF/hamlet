import { MigrationChecklistGenerator } from '../../src/core/MigrationChecklistGenerator.js';

describe('MigrationChecklistGenerator', () => {
  let generator;

  beforeEach(() => {
    generator = new MigrationChecklistGenerator();
  });

  const graph = { nodes: [], edges: new Map() };

  describe('generate', () => {
    it('should generate markdown with summary section', () => {
      const results = [
        { path: 'test.js', confidence: 95, warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('# Migration Checklist');
      expect(checklist).toContain('Total files');
    });

    it('should group files by confidence level', () => {
      const results = [
        { path: 'high.js', confidence: 95, warnings: [], todos: [], type: 'test' },
        { path: 'medium.js', confidence: 75, warnings: ['minor issue'], todos: [], type: 'test' },
        { path: 'low.js', confidence: 50, warnings: [], todos: ['manual fix'], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('Fully Converted');
      expect(checklist).toContain('Needs Review');
      expect(checklist).toContain('high.js');
      expect(checklist).toContain('medium.js');
    });

    it('should list TODOs and WARNINGs', () => {
      const results = [
        { path: 'test.js', confidence: 75, warnings: ['mock issue'], todos: ['update mocks'], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('WARNING: mock issue');
      expect(checklist).toContain('TODO: update mocks');
    });

    it('should handle empty results', () => {
      const checklist = generator.generate(graph, []);

      expect(checklist).toContain('# Migration Checklist');
      expect(checklist).toContain('Total files:** 0');
    });

    it('should include manual section for failed files', () => {
      const results = [
        { path: 'broken.js', confidence: 0, status: 'failed', error: 'parse error', warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('Manual Steps Required');
      expect(checklist).toContain('broken.js');
      expect(checklist).toContain('parse error');
    });

    it('should include config section for config files', () => {
      const results = [
        { path: 'jest.config.js', confidence: 90, warnings: [], todos: ['update paths'], type: 'config' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('Config Changes');
      expect(checklist).toContain('jest.config.js');
    });

    it('should mark fully converted files as checked', () => {
      const results = [
        { path: 'done.js', confidence: 95, warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('[x]');
    });

    it('should mark needs-review files as unchecked', () => {
      const results = [
        { path: 'review.js', confidence: 75, warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('[ ]');
    });

    it('should show confidence percentages', () => {
      const results = [
        { path: 'test.js', confidence: 87, warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('(87%)');
    });

    it('should count high, medium, low, and failed in summary', () => {
      const results = [
        { path: 'a.js', confidence: 95, warnings: [], todos: [], type: 'test' },
        { path: 'b.js', confidence: 80, warnings: [], todos: [], type: 'test' },
        { path: 'c.js', confidence: 50, warnings: [], todos: [], type: 'test' },
        { path: 'd.js', confidence: 0, status: 'failed', warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('High confidence (>=90%):** 1');
      expect(checklist).toContain('Medium confidence (70-89%):** 1');
      expect(checklist).toContain('Low confidence (<70%):** 1');
      expect(checklist).toContain('Failed/Manual:** 1');
    });

    it('should handle all-high-confidence scenario', () => {
      const results = [
        { path: 'a.js', confidence: 95, warnings: [], todos: [], type: 'test' },
        { path: 'b.js', confidence: 100, warnings: [], todos: [], type: 'test' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('Fully Converted');
      expect(checklist).not.toContain('Needs Review');
      expect(checklist).not.toContain('Manual Steps');
    });

    it('should handle mixed types correctly', () => {
      const results = [
        { path: 'test.js', confidence: 95, warnings: [], todos: [], type: 'test' },
        { path: 'jest.config.js', confidence: 80, warnings: [], todos: ['check paths'], type: 'config' },
        { path: 'helper.js', confidence: 0, status: 'failed', error: 'crash', warnings: [], todos: [], type: 'helper' },
      ];

      const checklist = generator.generate(graph, results);

      expect(checklist).toContain('Fully Converted');
      expect(checklist).toContain('Config Changes');
      expect(checklist).toContain('Manual Steps Required');
    });
  });
});
