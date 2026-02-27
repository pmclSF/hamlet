import path from 'path';
import { safePath } from '../../src/server/pathUtils.js';

describe('safePath', () => {
  const root = '/projects/myapp';

  describe('valid paths', () => {
    it('should resolve a relative path within root', () => {
      const result = safePath('src/index.js', root);
      expect(result).toBe(path.join(root, 'src/index.js'));
    });

    it('should resolve an absolute path within root', () => {
      const result = safePath('/projects/myapp/src/index.js', root);
      expect(result).toBe(path.join(root, 'src/index.js'));
    });

    it('should accept the root directory itself', () => {
      const result = safePath('.', root);
      expect(result).toBe(path.resolve(root));
    });

    it('should resolve a deeply nested path', () => {
      const result = safePath('a/b/c/d.js', root);
      expect(result).toBe(path.join(root, 'a/b/c/d.js'));
    });

    it('should resolve path with internal .. that stays within root', () => {
      const result = safePath('src/../lib/util.js', root);
      expect(result).toBe(path.join(root, 'lib/util.js'));
    });
  });

  describe('path traversal attacks', () => {
    it('should reject relative paths that escape the root', () => {
      expect(() => safePath('../../etc/passwd', root)).toThrow(
        'Path outside project root'
      );
    });

    it('should reject absolute paths outside the root', () => {
      expect(() => safePath('/etc/passwd', root)).toThrow(
        'Path outside project root'
      );
    });

    it('should reject paths that are prefix matches but not children', () => {
      // /projects/myapp vs /projects/myapplication — not a child
      expect(() => safePath('/projects/myapplication/foo', root)).toThrow(
        'Path outside project root'
      );
    });

    it('should reject deeply traversed paths', () => {
      expect(() => safePath('../../../../../../../etc/shadow', root)).toThrow(
        'Path outside project root'
      );
    });
  });

  describe('null byte injection', () => {
    it('should reject paths containing null bytes', () => {
      expect(() => safePath('src/foo\0.js', root)).toThrow(
        'Path contains null bytes'
      );
    });

    it('should reject null byte at start of path', () => {
      expect(() => safePath('\0/etc/passwd', root)).toThrow(
        'Path contains null bytes'
      );
    });
  });

  describe('invalid input', () => {
    it('should reject empty string', () => {
      expect(() => safePath('', root)).toThrow(
        'Path must be a non-empty string'
      );
    });

    it('should reject non-string input', () => {
      expect(() => safePath(null, root)).toThrow(
        'Path must be a non-empty string'
      );
    });

    it('should reject undefined input', () => {
      expect(() => safePath(undefined, root)).toThrow(
        'Path must be a non-empty string'
      );
    });
  });
});
