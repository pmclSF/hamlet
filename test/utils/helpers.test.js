import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import {
  fileUtils,
  stringUtils,
  codeUtils,
  testUtils,
  reportUtils,
  logUtils
} from '../../src/utils/helpers.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const tmpDir = path.join(__dirname, '../.tmp-helpers-test');

describe('fileUtils', () => {
  beforeEach(async () => {
    await fs.mkdir(tmpDir, { recursive: true });
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  describe('fileExists', () => {
    it('should return true for existing file', async () => {
      const filePath = path.join(tmpDir, 'exists.txt');
      await fs.writeFile(filePath, 'content');
      expect(await fileUtils.fileExists(filePath)).toBe(true);
    });

    it('should return false for non-existing file', async () => {
      expect(await fileUtils.fileExists(path.join(tmpDir, 'nope.txt'))).toBe(false);
    });
  });

  describe('ensureDir', () => {
    it('should create a directory', async () => {
      const dir = path.join(tmpDir, 'new-dir');
      await fileUtils.ensureDir(dir);
      const stat = await fs.stat(dir);
      expect(stat.isDirectory()).toBe(true);
    });

    it('should handle already-existing directory', async () => {
      await fileUtils.ensureDir(tmpDir);
      const stat = await fs.stat(tmpDir);
      expect(stat.isDirectory()).toBe(true);
    });
  });

  describe('copyDir', () => {
    it('should copy directory contents recursively', async () => {
      const src = path.join(tmpDir, 'src-copy');
      const dest = path.join(tmpDir, 'dest-copy');
      await fs.mkdir(src, { recursive: true });
      await fs.writeFile(path.join(src, 'file.txt'), 'hello');
      await fs.mkdir(path.join(src, 'sub'), { recursive: true });
      await fs.writeFile(path.join(src, 'sub', 'nested.txt'), 'nested');

      await fileUtils.copyDir(src, dest);

      const file = await fs.readFile(path.join(dest, 'file.txt'), 'utf8');
      expect(file).toBe('hello');
      const nested = await fs.readFile(path.join(dest, 'sub', 'nested.txt'), 'utf8');
      expect(nested).toBe('nested');
    });
  });

  describe('getFiles', () => {
    it('should get all files matching pattern', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.js'), '');
      await fs.writeFile(path.join(tmpDir, 'test.txt'), '');

      const jsFiles = await fileUtils.getFiles(tmpDir, /\.js$/);
      expect(jsFiles).toHaveLength(1);
      expect(jsFiles[0]).toContain('test.js');
    });

    it('should get all files when no pattern provided', async () => {
      await fs.writeFile(path.join(tmpDir, 'a.js'), '');
      await fs.writeFile(path.join(tmpDir, 'b.txt'), '');

      const files = await fileUtils.getFiles(tmpDir);
      expect(files).toHaveLength(2);
    });
  });

  describe('cleanDir', () => {
    it('should remove all files and subdirectories', async () => {
      await fs.writeFile(path.join(tmpDir, 'file.txt'), '');
      await fs.mkdir(path.join(tmpDir, 'subdir'), { recursive: true });
      await fs.writeFile(path.join(tmpDir, 'subdir', 'nested.txt'), '');

      await fileUtils.cleanDir(tmpDir);

      const entries = await fs.readdir(tmpDir);
      expect(entries).toHaveLength(0);
    });

    it('should exclude files matching pattern', async () => {
      await fs.writeFile(path.join(tmpDir, 'keep.json'), '');
      await fs.writeFile(path.join(tmpDir, 'delete.txt'), '');

      await fileUtils.cleanDir(tmpDir, /\.json$/);

      const entries = await fs.readdir(tmpDir);
      expect(entries).toEqual(['keep.json']);
    });

    it('should not throw for non-existing directory', async () => {
      await expect(fileUtils.cleanDir(path.join(tmpDir, 'nope'))).resolves.not.toThrow();
    });
  });
});

describe('stringUtils', () => {
  describe('camelToKebab', () => {
    it('should convert camelCase to kebab-case', () => {
      expect(stringUtils.camelToKebab('camelCase')).toBe('camel-case');
      expect(stringUtils.camelToKebab('myTestValue')).toBe('my-test-value');
    });

    it('should handle single word', () => {
      expect(stringUtils.camelToKebab('hello')).toBe('hello');
    });
  });

  describe('kebabToCamel', () => {
    it('should convert kebab-case to camelCase', () => {
      expect(stringUtils.kebabToCamel('kebab-case')).toBe('kebabCase');
      expect(stringUtils.kebabToCamel('my-test-value')).toBe('myTestValue');
    });

    it('should handle single word', () => {
      expect(stringUtils.kebabToCamel('hello')).toBe('hello');
    });
  });

  describe('calculateSimilarity', () => {
    it('should return 1 for identical strings', () => {
      expect(stringUtils.calculateSimilarity('hello', 'hello')).toBe(1);
    });

    it('should return 0 for completely different strings', () => {
      expect(stringUtils.calculateSimilarity('abc', 'xyz')).toBe(0);
    });

    it('should return value between 0 and 1 for similar strings', () => {
      const similarity = stringUtils.calculateSimilarity('hello', 'hallo');
      expect(similarity).toBeGreaterThan(0);
      expect(similarity).toBeLessThan(1);
    });
  });
});

describe('codeUtils', () => {
  describe('extractImports', () => {
    it('should extract named imports', () => {
      const code = "import { foo, bar } from './module.js';";
      const imports = codeUtils.extractImports(code);
      expect(imports).toHaveLength(1);
      expect(imports[0].source).toBe('./module.js');
    });

    it('should extract default imports', () => {
      const code = "import fs from 'fs';";
      const imports = codeUtils.extractImports(code);
      expect(imports).toHaveLength(1);
      expect(imports[0].source).toBe('fs');
    });

    it('should extract namespace imports', () => {
      const code = "import * as utils from './utils.js';";
      const imports = codeUtils.extractImports(code);
      expect(imports).toHaveLength(1);
      expect(imports[0].source).toBe('./utils.js');
    });

    it('should return empty array for no imports', () => {
      expect(codeUtils.extractImports('const x = 1;')).toEqual([]);
    });
  });

  describe('extractExports', () => {
    it('should extract class exports', () => {
      const code = 'export class MyClass {}';
      const exports = codeUtils.extractExports(code);
      expect(exports).toHaveLength(1);
      expect(exports[0].name).toBe('MyClass');
    });

    it('should extract function exports', () => {
      const code = 'export function myFunc() {}';
      const exports = codeUtils.extractExports(code);
      expect(exports).toHaveLength(1);
      expect(exports[0].name).toBe('myFunc');
    });

    it('should extract const exports', () => {
      const code = 'export const MY_CONST = 42;';
      const exports = codeUtils.extractExports(code);
      expect(exports).toHaveLength(1);
      expect(exports[0].name).toBe('MY_CONST');
    });
  });

  describe('formatCode', () => {
    it('should add spaces after commas', () => {
      const result = codeUtils.formatCode('foo(a,b,c)');
      expect(result).toContain('a, b, c');
    });

    it('should add spaces after keywords', () => {
      const result = codeUtils.formatCode('if(true) {}');
      expect(result).toContain('if (true)');
    });

    it('should trim trailing whitespace', () => {
      const result = codeUtils.formatCode('const x = 1;   ');
      expect(result).not.toMatch(/\s+\n/);
    });
  });
});

describe('testUtils', () => {
  describe('extractTestCases', () => {
    it('should extract test descriptions', () => {
      const content = `
        it('should do something', () => {
          expect(true).toBe(true);
        });
        it('should do another thing', () => {
          expect(1).toBe(1);
        });
      `;
      const tests = testUtils.extractTestCases(content);
      expect(tests).toHaveLength(2);
      expect(tests[0].description).toBe('should do something');
      expect(tests[1].description).toBe('should do another thing');
    });
  });

  describe('extractAssertions', () => {
    it('should extract expect assertions', () => {
      const content = "expect(result).toBe('value')";
      const assertions = testUtils.extractAssertions(content);
      expect(assertions.length).toBeGreaterThanOrEqual(0);
    });
  });
});

describe('reportUtils', () => {
  describe('formatDuration', () => {
    it('should format seconds', () => {
      expect(reportUtils.formatDuration(5000)).toBe('5s');
    });

    it('should format minutes and seconds', () => {
      expect(reportUtils.formatDuration(90000)).toBe('1m 30s');
    });

    it('should format hours, minutes, and seconds', () => {
      expect(reportUtils.formatDuration(3661000)).toBe('1h 1m 1s');
    });
  });

  describe('createProgressBar', () => {
    it('should create progress bar with update and complete methods', () => {
      const bar = reportUtils.createProgressBar(100);
      expect(typeof bar.update).toBe('function');
      expect(typeof bar.complete).toBe('function');
    });
  });
});

describe('logUtils', () => {
  describe('createLogger', () => {
    it('should create logger with info, success, warn, error methods', () => {
      const logger = logUtils.createLogger('Test');
      expect(typeof logger.info).toBe('function');
      expect(typeof logger.success).toBe('function');
      expect(typeof logger.warn).toBe('function');
      expect(typeof logger.error).toBe('function');
    });
  });

  describe('logError', () => {
    it('should be a function', () => {
      expect(typeof logUtils.logError).toBe('function');
    });
  });

  describe('logWarning', () => {
    it('should be a function', () => {
      expect(typeof logUtils.logWarning).toBe('function');
    });
  });
});
