import { ImportRewriter } from '../../src/core/ImportRewriter.js';

describe('ImportRewriter', () => {
  let rewriter;

  beforeEach(() => {
    rewriter = new ImportRewriter();
  });

  describe('rewrite', () => {
    it('should rewrite ES named import path', () => {
      const content = `import { foo } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import { foo } from './new.js';`);
    });

    it('should rewrite ES default import path', () => {
      const content = `import Foo from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import Foo from './new.js';`);
    });

    it('should rewrite ES namespace import path', () => {
      const content = `import * as utils from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import * as utils from './new.js';`);
    });

    it('should rewrite ES mixed import path', () => {
      const content = `import Foo, { bar } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import Foo, { bar } from './new.js';`);
    });

    it('should rewrite require path', () => {
      const content = `const foo = require('./old.js');`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`const foo = require('./new.js');`);
    });

    it('should rewrite destructured require path', () => {
      const content = `const { foo, bar } = require('./old.js');`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`const { foo, bar } = require('./new.js');`);
    });

    it('should rewrite dynamic import path', () => {
      const content = `const mod = await import('./old.js');`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain("'./new.js'");
    });

    it('should rewrite re-export path', () => {
      const content = `export { foo } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`export { foo } from './new.js';`);
    });

    it('should rewrite export-all path', () => {
      const content = `export * from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`export * from './new.js';`);
    });

    it('should rewrite type-only import path', () => {
      const content = `import type { Foo } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import type { Foo } from './new.js';`);
    });

    it('should rewrite side-effect import path', () => {
      const content = `import './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import './new.js';`);
    });

    it('should handle path with extension correctly', () => {
      const content = `import { x } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      expect(rewriter.rewrite(content, renames)).toBe(`import { x } from './new.js';`);
    });

    it('should handle path without extension', () => {
      const content = `import { x } from './old';`;
      const renames = new Map([['./old', './new']]);

      expect(rewriter.rewrite(content, renames)).toBe(`import { x } from './new';`);
    });

    it('should handle path with ../ traversal', () => {
      const content = `import { x } from '../utils/old.js';`;
      const renames = new Map([['../utils/old.js', '../utils/new.js']]);

      expect(rewriter.rewrite(content, renames)).toBe(`import { x } from '../utils/new.js';`);
    });

    it('should rewrite multiple imports from same module on different lines', () => {
      const content = `import { foo } from './old.js';\nimport { bar } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import { foo } from './new.js';\nimport { bar } from './new.js';`);
    });

    it('should NOT rewrite substring matches', () => {
      const content = `import { foo } from './bar-utils.js';`;
      const renames = new Map([['./bar.js', './baz.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import { foo } from './bar-utils.js';`);
    });

    it('should NOT rewrite imports not in rename map', () => {
      const content = `import { foo } from './untouched.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import { foo } from './untouched.js';`);
    });

    it('should NEVER rewrite node_modules imports', () => {
      const content = `import React from 'react';`;
      const renames = new Map([['react', 'preact']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import React from 'react';`);
    });

    it('should return unchanged content when no rewritable imports', () => {
      const content = `const x = 1;\nconsole.log(x);`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(content);
    });

    it('should NOT rewrite import inside a single-line comment', () => {
      const content = `// import { foo } from './old.js';\nimport { bar } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain("// import { foo } from './old.js';");
      expect(result).toContain("import { bar } from './new.js';");
    });

    it('should NOT rewrite import inside a block comment', () => {
      const content = `/* import { foo } from './old.js'; */\nimport { bar } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain("/* import { foo } from './old.js'; */");
      expect(result).toContain("import { bar } from './new.js';");
    });

    it('should handle multiline import statement', () => {
      const content = `import {\n  foo,\n  bar,\n  baz\n} from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain("from './new.js'");
      expect(result).not.toContain("from './old.js'");
    });

    it('should preserve import with trailing comment', () => {
      const content = `import { foo } from './old.js'; // important`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain("'./new.js'");
      expect(result).toContain('// important');
    });

    it('should handle aliased import', () => {
      const content = `import { foo as myFoo } from './old.js';`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import { foo as myFoo } from './new.js';`);
    });

    it('should handle empty rename map', () => {
      const content = `import { foo } from './old.js';`;
      const renames = new Map();

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(content);
    });

    it('should handle content with no imports at all', () => {
      const content = `const x = 1;\nconst y = 2;`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(content);
    });

    it('should handle double-quoted imports', () => {
      const content = `import { foo } from "./old.js";`;
      const renames = new Map([['./old.js', './new.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toBe(`import { foo } from "./new.js";`);
    });

    it('should match extensionless rename map against import with extension', () => {
      const content = `import { foo } from './helpers.js';`;
      const renames = new Map([['./helpers', './utils']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain('./utils');
    });

    it('should match import without extension against rename map with extension', () => {
      const content = `import { foo } from './helpers';`;
      const renames = new Map([['./helpers.js', './utils.js']]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain('./utils');
    });

    it('should handle integration: converted helper + import rewrite', () => {
      // Simulate a scenario where a helper was renamed
      const content = [
        `import { createMock } from './helpers/testHelper.js';`,
        `import { expect } from 'vitest';`,
        '',
        'describe("test", () => {',
        '  it("works", () => {',
        '    const mock = createMock();',
        '    expect(mock).toBeDefined();',
        '  });',
        '});',
      ].join('\n');

      const renames = new Map([
        ['./helpers/testHelper.js', './helpers/testHelper.spec.js'],
      ]);

      const result = rewriter.rewrite(content, renames);

      expect(result).toContain("from './helpers/testHelper.spec.js'");
      expect(result).toContain("from 'vitest'"); // Node module untouched
    });
  });
});
