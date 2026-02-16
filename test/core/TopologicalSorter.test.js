import { TopologicalSorter } from '../../src/core/TopologicalSorter.js';

describe('TopologicalSorter', () => {
  let sorter;

  beforeEach(() => {
    sorter = new TopologicalSorter();
  });

  describe('sort', () => {
    it('should sort a simple DAG (leaves first)', () => {
      // A depends on B, B depends on C
      const graph = {
        nodes: ['A', 'B', 'C'],
        edges: new Map([
          ['A', ['B']],
          ['B', ['C']],
          ['C', []],
        ]),
      };

      const sorted = sorter.sort(graph);

      // C should come before B, B before A
      expect(sorted.indexOf('C')).toBeLessThan(sorted.indexOf('B'));
      expect(sorted.indexOf('B')).toBeLessThan(sorted.indexOf('A'));
    });

    it('should handle multiple roots', () => {
      // A and B both depend on C
      const graph = {
        nodes: ['A', 'B', 'C'],
        edges: new Map([
          ['A', ['C']],
          ['B', ['C']],
          ['C', []],
        ]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted.indexOf('C')).toBeLessThan(sorted.indexOf('A'));
      expect(sorted.indexOf('C')).toBeLessThan(sorted.indexOf('B'));
    });

    it('should handle multiple sinks (leaves)', () => {
      // A depends on both B and C (B and C are leaves)
      const graph = {
        nodes: ['A', 'B', 'C'],
        edges: new Map([
          ['A', ['B', 'C']],
          ['B', []],
          ['C', []],
        ]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted.indexOf('B')).toBeLessThan(sorted.indexOf('A'));
      expect(sorted.indexOf('C')).toBeLessThan(sorted.indexOf('A'));
    });

    it('should handle cycles without infinite looping', () => {
      // A → B → A (cycle)
      const graph = {
        nodes: ['A', 'B'],
        edges: new Map([
          ['A', ['B']],
          ['B', ['A']],
        ]),
      };

      const sorted = sorter.sort(graph);

      // Should return all nodes even with cycle
      expect(sorted).toHaveLength(2);
      expect(sorted).toContain('A');
      expect(sorted).toContain('B');
    });

    it('should return empty array for empty graph', () => {
      const graph = {
        nodes: [],
        edges: new Map(),
      };

      const sorted = sorter.sort(graph);

      expect(sorted).toEqual([]);
    });

    it('should handle single node graph', () => {
      const graph = {
        nodes: ['A'],
        edges: new Map([['A', []]]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted).toEqual(['A']);
    });

    it('should handle diamond dependency correctly', () => {
      // A depends on B and C, both depend on D
      const graph = {
        nodes: ['A', 'B', 'C', 'D'],
        edges: new Map([
          ['A', ['B', 'C']],
          ['B', ['D']],
          ['C', ['D']],
          ['D', []],
        ]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted.indexOf('D')).toBeLessThan(sorted.indexOf('B'));
      expect(sorted.indexOf('D')).toBeLessThan(sorted.indexOf('C'));
      expect(sorted.indexOf('B')).toBeLessThan(sorted.indexOf('A'));
      expect(sorted.indexOf('C')).toBeLessThan(sorted.indexOf('A'));
    });

    it('should handle independent subgraphs', () => {
      // Two disconnected pairs
      const graph = {
        nodes: ['A', 'B', 'C', 'D'],
        edges: new Map([
          ['A', ['B']],
          ['B', []],
          ['C', ['D']],
          ['D', []],
        ]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted).toHaveLength(4);
      expect(sorted.indexOf('B')).toBeLessThan(sorted.indexOf('A'));
      expect(sorted.indexOf('D')).toBeLessThan(sorted.indexOf('C'));
    });

    it('should place independent nodes first (no dependencies)', () => {
      const graph = {
        nodes: ['A', 'B', 'C'],
        edges: new Map([
          ['A', []],
          ['B', []],
          ['C', ['A']],
        ]),
      };

      const sorted = sorter.sort(graph);

      // A and B have no deps, should come before C
      expect(sorted.indexOf('A')).toBeLessThan(sorted.indexOf('C'));
    });

    it('should handle long chain', () => {
      const nodes = ['A', 'B', 'C', 'D', 'E'];
      const edges = new Map([
        ['A', ['B']],
        ['B', ['C']],
        ['C', ['D']],
        ['D', ['E']],
        ['E', []],
      ]);

      const sorted = sorter.sort({ nodes, edges });

      for (let i = 0; i < sorted.length - 1; i++) {
        // Each node should come after its dependency
        const nodeIdx = nodes.indexOf(sorted[i]);
        const nextIdx = nodes.indexOf(sorted[i + 1]);
        // In original ordering, higher index means closer to root
        // In sorted output, lower index means earlier (leaf)
      }
      expect(sorted[0]).toBe('E');
      expect(sorted[sorted.length - 1]).toBe('A');
    });

    it('should handle three-node cycle and still return all nodes', () => {
      const graph = {
        nodes: ['A', 'B', 'C'],
        edges: new Map([
          ['A', ['B']],
          ['B', ['C']],
          ['C', ['A']],
        ]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted).toHaveLength(3);
      expect(sorted).toContain('A');
      expect(sorted).toContain('B');
      expect(sorted).toContain('C');
    });

    it('should handle mixed cycle and non-cycle nodes', () => {
      // D depends on nothing, A→B→A is a cycle, C depends on D
      const graph = {
        nodes: ['A', 'B', 'C', 'D'],
        edges: new Map([
          ['A', ['B']],
          ['B', ['A']],
          ['C', ['D']],
          ['D', []],
        ]),
      };

      const sorted = sorter.sort(graph);

      expect(sorted).toHaveLength(4);
      // D should come before C
      expect(sorted.indexOf('D')).toBeLessThan(sorted.indexOf('C'));
    });
  });
});
