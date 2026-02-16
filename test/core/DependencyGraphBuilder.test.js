import path from 'path';
import { DependencyGraphBuilder } from '../../src/core/DependencyGraphBuilder.js';

describe('DependencyGraphBuilder', () => {
  let builder;
  const root = '/project';

  beforeEach(() => {
    builder = new DependencyGraphBuilder();
  });

  function makeFile(relativePath, content) {
    return {
      path: path.join(root, relativePath),
      relativePath,
      content,
    };
  }

  describe('build', () => {
    it('should extract ES named imports', () => {
      const files = [
        makeFile('a.js', `import { foo } from './b.js';`),
        makeFile('b.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract ES default imports', () => {
      const files = [
        makeFile('a.js', `import foo from './b.js';`),
        makeFile('b.js', `export default 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract ES namespace imports', () => {
      const files = [
        makeFile('a.js', `import * as utils from './b.js';`),
        makeFile('b.js', `export const x = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract require calls', () => {
      const files = [
        makeFile('a.js', `const foo = require('./b.js');`),
        makeFile('b.js', `module.exports = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract destructured require calls', () => {
      const files = [
        makeFile('a.js', `const { foo, bar } = require('./b.js');`),
        makeFile('b.js', `module.exports = { foo: 1, bar: 2 };`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract dynamic imports', () => {
      const files = [
        makeFile('a.js', `const foo = await import('./b.js');`),
        makeFile('b.js', `export default 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract re-exports', () => {
      const files = [
        makeFile('a.js', `export { foo } from './b.js';`),
        makeFile('b.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should extract side-effect imports', () => {
      const files = [
        makeFile('a.js', `import './setup.js';`),
        makeFile('setup.js', `global.x = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'setup.js'));
    });

    it('should resolve relative paths without extensions', () => {
      const files = [
        makeFile('a.js', `import { foo } from './b';`),
        makeFile('b.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should resolve index file imports', () => {
      const files = [
        makeFile('a.js', `import { foo } from './helpers';`),
        makeFile('helpers/index.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toContain(path.join(root, 'helpers', 'index.js'));
    });

    it('should resolve ../ traversal paths', () => {
      const files = [
        makeFile('sub/a.js', `import { foo } from '../b.js';`),
        makeFile('b.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'sub', 'a.js'))).toContain(path.join(root, 'b.js'));
    });

    it('should handle multiple imports from same file', () => {
      const files = [
        makeFile('a.js', `import { foo } from './b.js';\nimport { bar } from './b.js';`),
        makeFile('b.js', `export const foo = 1;\nexport const bar = 2;`),
      ];

      const graph = builder.build(files);

      const deps = graph.edges.get(path.join(root, 'a.js'));
      expect(deps).toContain(path.join(root, 'b.js'));
      // Should not have duplicates
      expect(deps.filter(d => d === path.join(root, 'b.js'))).toHaveLength(1);
    });

    it('should NOT add node_modules imports to graph', () => {
      const files = [
        makeFile('a.js', `import React from 'react';\nimport { foo } from './b.js';`),
        makeFile('b.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      const deps = graph.edges.get(path.join(root, 'a.js'));
      expect(deps).toHaveLength(1);
      expect(deps).toContain(path.join(root, 'b.js'));
    });

    it('should handle import from nonexistent file without crashing', () => {
      const files = [
        makeFile('a.js', `import { foo } from './nonexistent.js';`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toHaveLength(0);
      expect(graph.warnings.length).toBeGreaterThan(0);
    });

    it('should handle file with zero imports (leaf node)', () => {
      const files = [
        makeFile('leaf.js', `export const x = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'leaf.js'))).toHaveLength(0);
    });

    it('should handle file with many imports (high fan-out)', () => {
      const importLines = Array.from({ length: 20 }, (_, i) =>
        `import { x${i} } from './m${i}.js';`
      ).join('\n');

      const modules = Array.from({ length: 20 }, (_, i) =>
        makeFile(`m${i}.js`, `export const x${i} = ${i};`)
      );

      const files = [makeFile('main.js', importLines), ...modules];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'main.js'))).toHaveLength(20);
    });

    // Circular dependency detection
    it('should detect A→B→A circular dependency', () => {
      const files = [
        makeFile('a.js', `import { b } from './b.js';`),
        makeFile('b.js', `import { a } from './a.js';`),
      ];

      const graph = builder.build(files);

      expect(graph.cycles.length).toBeGreaterThan(0);
    });

    it('should detect A→B→C→A circular dependency', () => {
      const files = [
        makeFile('a.js', `import { b } from './b.js';`),
        makeFile('b.js', `import { c } from './c.js';`),
        makeFile('c.js', `import { a } from './a.js';`),
      ];

      const graph = builder.build(files);

      expect(graph.cycles.length).toBeGreaterThan(0);
    });

    it('should handle diamond dependency without false cycle', () => {
      const files = [
        makeFile('a.js', `import { b } from './b.js';\nimport { c } from './c.js';`),
        makeFile('b.js', `import { d } from './d.js';`),
        makeFile('c.js', `import { d } from './d.js';`),
        makeFile('d.js', `export const d = 1;`),
      ];

      const graph = builder.build(files);

      // Diamond is not a cycle
      expect(graph.cycles).toHaveLength(0);
    });

    it('should handle large graph with 15+ files', () => {
      const files = [];
      // Create a chain: f0 → f1 → f2 → ... → f14
      for (let i = 0; i < 15; i++) {
        const content = i < 14
          ? `import { x } from './f${i + 1}.js';\nexport const y${i} = 1;`
          : `export const x = 'leaf';`;
        files.push(makeFile(`f${i}.js`, content));
      }

      const graph = builder.build(files);

      expect(graph.nodes).toHaveLength(15);
      expect(graph.cycles).toHaveLength(0);
    });

    it('should not match imports inside comments', () => {
      const files = [
        makeFile('a.js', `// import { foo } from './b.js';\nexport const x = 1;`),
        makeFile('b.js', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.js'))).toHaveLength(0);
    });

    it('should extract export-all re-exports', () => {
      const files = [
        makeFile('index.js', `export * from './utils.js';`),
        makeFile('utils.js', `export const x = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'index.js'))).toContain(path.join(root, 'utils.js'));
    });

    it('should extract type-only imports', () => {
      const files = [
        makeFile('a.ts', `import type { Foo } from './types.ts';`),
        makeFile('types.ts', `export type Foo = string;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.ts'))).toContain(path.join(root, 'types.ts'));
    });

    it('should resolve TypeScript files without extension', () => {
      const files = [
        makeFile('a.ts', `import { foo } from './b';`),
        makeFile('b.ts', `export const foo = 1;`),
      ];

      const graph = builder.build(files);

      expect(graph.edges.get(path.join(root, 'a.ts'))).toContain(path.join(root, 'b.ts'));
    });
  });
});
