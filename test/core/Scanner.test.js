import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { Scanner } from '../../src/core/Scanner.js';

describe('Scanner', () => {
  let scanner;
  let tmpDir;

  beforeEach(async () => {
    scanner = new Scanner();
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-scanner-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  describe('scan', () => {
    it('should scan a directory and return file metadata', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.js'), 'console.log("hello")');
      await fs.writeFile(path.join(tmpDir, 'test2.ts'), 'const x: number = 1');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(2);
      expect(results[0]).toHaveProperty('path');
      expect(results[0]).toHaveProperty('relativePath');
      expect(results[0]).toHaveProperty('size');
      expect(results[0].size).toBeGreaterThan(0);
    });

    it('should return relative paths from root', async () => {
      const subdir = path.join(tmpDir, 'src');
      await fs.mkdir(subdir);
      await fs.writeFile(path.join(subdir, 'index.js'), 'export default {}');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe(path.join('src', 'index.js'));
    });

    it('should ignore node_modules by default', async () => {
      const nmDir = path.join(tmpDir, 'node_modules');
      await fs.mkdir(nmDir);
      await fs.writeFile(path.join(nmDir, 'pkg.js'), 'module.exports = {}');
      await fs.writeFile(path.join(tmpDir, 'app.js'), 'import "./pkg"');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe('app.js');
    });

    it('should ignore .git directory by default', async () => {
      const gitDir = path.join(tmpDir, '.git');
      await fs.mkdir(gitDir);
      await fs.writeFile(path.join(gitDir, 'config'), '[core]');
      await fs.writeFile(path.join(tmpDir, 'app.js'), 'export default 1');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
    });

    it('should ignore dist directory by default', async () => {
      const distDir = path.join(tmpDir, 'dist');
      await fs.mkdir(distDir);
      await fs.writeFile(path.join(distDir, 'bundle.js'), 'var x = 1');
      await fs.writeFile(path.join(tmpDir, 'src.js'), 'export const x = 1');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe('src.js');
    });

    it('should ignore coverage directory by default', async () => {
      const covDir = path.join(tmpDir, 'coverage');
      await fs.mkdir(covDir);
      await fs.writeFile(path.join(covDir, 'lcov.info'), '');
      await fs.writeFile(path.join(tmpDir, 'test.js'), 'test()');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
    });

    it('should handle empty directories', async () => {
      const results = await scanner.scan(tmpDir);
      expect(results).toHaveLength(0);
    });

    it('should handle nested empty directories', async () => {
      await fs.mkdir(path.join(tmpDir, 'a', 'b', 'c'), { recursive: true });
      const results = await scanner.scan(tmpDir);
      expect(results).toHaveLength(0);
    });

    it('should scan nested directories recursively', async () => {
      await fs.mkdir(path.join(tmpDir, 'a', 'b'), { recursive: true });
      await fs.writeFile(path.join(tmpDir, 'a', 'one.js'), '1');
      await fs.writeFile(path.join(tmpDir, 'a', 'b', 'two.js'), '2');
      await fs.writeFile(path.join(tmpDir, 'root.js'), '0');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(3);
      const relativePaths = results.map(r => r.relativePath).sort();
      expect(relativePaths).toContain('root.js');
      expect(relativePaths).toContain(path.join('a', 'one.js'));
      expect(relativePaths).toContain(path.join('a', 'b', 'two.js'));
    });

    it('should support configurable include globs', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.js'), 'test');
      await fs.writeFile(path.join(tmpDir, 'test.ts'), 'test');
      await fs.writeFile(path.join(tmpDir, 'data.json'), '{}');

      const results = await scanner.scan(tmpDir, { include: ['*.js'] });

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe('test.js');
    });

    it('should support configurable exclude globs', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.js'), 'test');
      await fs.writeFile(path.join(tmpDir, 'test.min.js'), 'minified');
      await fs.writeFile(path.join(tmpDir, 'data.json'), '{}');

      const results = await scanner.scan(tmpDir, { exclude: ['*.json'] });

      expect(results).toHaveLength(2);
      const names = results.map(r => r.relativePath).sort();
      expect(names).toEqual(['test.js', 'test.min.js']);
    });

    it('should support additional ignore patterns', async () => {
      const myDir = path.join(tmpDir, 'vendor');
      await fs.mkdir(myDir);
      await fs.writeFile(path.join(myDir, 'lib.js'), 'lib');
      await fs.writeFile(path.join(tmpDir, 'app.js'), 'app');

      const results = await scanner.scan(tmpDir, { ignore: ['vendor'] });

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe('app.js');
    });

    it('should handle permission errors gracefully', async () => {
      // Create an unreadable directory
      const restricted = path.join(tmpDir, 'restricted');
      await fs.mkdir(restricted);
      await fs.writeFile(path.join(restricted, 'secret.js'), 'secret');
      await fs.chmod(restricted, 0o000);

      await fs.writeFile(path.join(tmpDir, 'public.js'), 'public');

      const results = await scanner.scan(tmpDir);

      // Should still find the public file without throwing
      expect(results.some(r => r.relativePath === 'public.js')).toBe(true);

      // Restore permissions for cleanup
      await fs.chmod(restricted, 0o755);
    });

    it('should include file size in results', async () => {
      const content = 'a'.repeat(100);
      await fs.writeFile(path.join(tmpDir, 'sized.js'), content);

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
      expect(results[0].size).toBe(100);
    });

    it('should return absolute paths in path property', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.js'), 'x');

      const results = await scanner.scan(tmpDir);

      expect(path.isAbsolute(results[0].path)).toBe(true);
    });

    it('should combine include and exclude filters', async () => {
      await fs.writeFile(path.join(tmpDir, 'app.js'), 'app');
      await fs.writeFile(path.join(tmpDir, 'app.test.js'), 'test');
      await fs.writeFile(path.join(tmpDir, 'app.ts'), 'ts');

      const results = await scanner.scan(tmpDir, {
        include: ['*.js'],
        exclude: ['*.test.js'],
      });

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe('app.js');
    });

    it('should handle .hamlet directory in ignore', async () => {
      const hamletDir = path.join(tmpDir, '.hamlet');
      await fs.mkdir(hamletDir);
      await fs.writeFile(path.join(hamletDir, 'state.json'), '{}');
      await fs.writeFile(path.join(tmpDir, 'test.js'), 'test');

      const results = await scanner.scan(tmpDir);

      expect(results).toHaveLength(1);
      expect(results[0].relativePath).toBe('test.js');
    });
  });
});
