import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { safePath } from '../../src/server/pathUtils.js';

describe('safePath', () => {
  const root = '/projects/myapp';

  describe('valid paths', () => {
    it('should resolve a relative path within root', async () => {
      const result = await safePath('src/index.js', root);
      expect(result).toBe(path.join(root, 'src/index.js'));
    });

    it('should resolve an absolute path within root', async () => {
      const result = await safePath('/projects/myapp/src/index.js', root);
      expect(result).toBe(path.join(root, 'src/index.js'));
    });

    it('should accept the root directory itself', async () => {
      const result = await safePath('.', root);
      expect(result).toBe(path.resolve(root));
    });

    it('should resolve a deeply nested path', async () => {
      const result = await safePath('a/b/c/d.js', root);
      expect(result).toBe(path.join(root, 'a/b/c/d.js'));
    });

    it('should resolve path with internal .. that stays within root', async () => {
      const result = await safePath('src/../lib/util.js', root);
      expect(result).toBe(path.join(root, 'lib/util.js'));
    });
  });

  describe('path traversal attacks', () => {
    it('should reject relative paths that escape the root', async () => {
      await expect(safePath('../../etc/passwd', root)).rejects.toThrow(
        'Path outside project root'
      );
    });

    it('should reject absolute paths outside the root', async () => {
      await expect(safePath('/etc/passwd', root)).rejects.toThrow(
        'Path outside project root'
      );
    });

    it('should reject paths that are prefix matches but not children', async () => {
      // /projects/myapp vs /projects/myapplication — not a child
      await expect(
        safePath('/projects/myapplication/foo', root)
      ).rejects.toThrow('Path outside project root');
    });

    it('should reject deeply traversed paths', async () => {
      await expect(
        safePath('../../../../../../../etc/shadow', root)
      ).rejects.toThrow('Path outside project root');
    });
  });

  describe('null byte injection', () => {
    it('should reject paths containing null bytes', async () => {
      await expect(safePath('src/foo\0.js', root)).rejects.toThrow(
        'Path contains null bytes'
      );
    });

    it('should reject null byte at start of path', async () => {
      await expect(safePath('\0/etc/passwd', root)).rejects.toThrow(
        'Path contains null bytes'
      );
    });
  });

  describe('invalid input', () => {
    it('should reject empty string', async () => {
      await expect(safePath('', root)).rejects.toThrow(
        'Path must be a non-empty string'
      );
    });

    it('should reject non-string input', async () => {
      await expect(safePath(null, root)).rejects.toThrow(
        'Path must be a non-empty string'
      );
    });

    it('should reject undefined input', async () => {
      await expect(safePath(undefined, root)).rejects.toThrow(
        'Path must be a non-empty string'
      );
    });
  });

  describe('symlink escape prevention', () => {
    let tmpDir;
    let projectRoot;

    beforeEach(async () => {
      tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'safepath-test-'));
      projectRoot = path.join(tmpDir, 'project');
      await fs.mkdir(projectRoot, { recursive: true });
    });

    afterEach(async () => {
      await fs.rm(tmpDir, { recursive: true, force: true });
    });

    it('should allow a real file inside the root', async () => {
      const filePath = path.join(projectRoot, 'file.txt');
      await fs.writeFile(filePath, 'ok');

      const result = await safePath('file.txt', projectRoot);
      expect(result).toBe(filePath);
    });

    it('should reject a symlink that points outside the root', async () => {
      // Create a target file outside the project root
      const outsideFile = path.join(tmpDir, 'secret.txt');
      await fs.writeFile(outsideFile, 'secret');

      // Create a symlink inside root that points to the outside file
      const linkPath = path.join(projectRoot, 'link.txt');
      await fs.symlink(outsideFile, linkPath);

      await expect(safePath('link.txt', projectRoot)).rejects.toThrow(
        'Path outside project root'
      );
    });

    it('should reject a symlinked directory that escapes root', async () => {
      // Create a directory outside root
      const outsideDir = path.join(tmpDir, 'outside');
      await fs.mkdir(outsideDir);
      await fs.writeFile(path.join(outsideDir, 'data.txt'), 'data');

      // Create a symlink inside root to the outside directory
      const linkDir = path.join(projectRoot, 'linked-dir');
      await fs.symlink(outsideDir, linkDir);

      await expect(
        safePath('linked-dir/data.txt', projectRoot)
      ).rejects.toThrow('Path outside project root');
    });

    it('should allow a symlink that stays within the root', async () => {
      // Create a real file inside root
      const realFile = path.join(projectRoot, 'real.txt');
      await fs.writeFile(realFile, 'data');

      // Create a symlink inside root that points to another file inside root
      const linkPath = path.join(projectRoot, 'alias.txt');
      await fs.symlink(realFile, linkPath);

      const result = await safePath('alias.txt', projectRoot);
      expect(result).toBe(linkPath);
    });
  });
});
