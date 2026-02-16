import { ErrorRecovery } from '../../src/core/ErrorRecovery.js';

describe('ErrorRecovery', () => {
  let recovery;

  beforeEach(() => {
    recovery = new ErrorRecovery();
  });

  describe('wrap', () => {
    it('should return result for successful function', () => {
      const fn = (x) => x * 2;
      const wrapped = recovery.wrap(fn);

      const { result, error } = wrapped(5);

      expect(result).toBe(10);
      expect(error).toBeNull();
    });

    it('should catch error and return fallback for throwing function', () => {
      const fn = () => { throw new Error('boom'); };
      const wrapped = recovery.wrap(fn, 'fallback');

      const { result, error } = wrapped();

      expect(result).toBe('fallback');
      expect(error).toBeInstanceOf(Error);
      expect(error.message).toBe('boom');
    });

    it('should return null as default fallback', () => {
      const fn = () => { throw new Error('boom'); };
      const wrapped = recovery.wrap(fn);

      const { result, error } = wrapped();

      expect(result).toBeNull();
      expect(error).toBeTruthy();
    });

    it('should handle async functions', async () => {
      const fn = async (x) => x + 1;
      const wrapped = recovery.wrap(fn);

      const { result, error } = await wrapped(5);

      expect(result).toBe(6);
      expect(error).toBeNull();
    });

    it('should catch async errors', async () => {
      const fn = async () => { throw new Error('async boom'); };
      const wrapped = recovery.wrap(fn, 'async-fallback');

      const { result, error } = await wrapped();

      expect(result).toBe('async-fallback');
      expect(error.message).toBe('async boom');
    });
  });

  describe('recoverFromParseError', () => {
    it('should process lines individually when full parse fails', () => {
      const content = 'line1\nline2\nline3';
      const error = new Error('parse failed');
      const processor = (line) => line.toUpperCase();

      const { recovered, warnings } = recovery.recoverFromParseError(content, error, processor);

      expect(recovered).toContain('LINE1');
      expect(recovered).toContain('LINE2');
      expect(recovered).toContain('LINE3');
    });

    it('should add HAMLET-WARNING for lines that fail individually', () => {
      const content = 'good\nbad\ngood2';
      const error = new Error('parse failed');
      const processor = (line) => {
        if (line === 'bad') throw new Error('bad line');
        return line.toUpperCase();
      };

      const { recovered, warnings } = recovery.recoverFromParseError(content, error, processor);

      expect(recovered).toContain('HAMLET-WARNING');
      expect(recovered).toContain('bad');
      expect(warnings.length).toBeGreaterThan(0);
    });

    it('should add top-level warning when all lines succeed', () => {
      const content = 'line1\nline2';
      const error = new Error('original parse error');
      const processor = (line) => line;

      const { recovered, warnings } = recovery.recoverFromParseError(content, error, processor);

      expect(recovered).toContain('HAMLET-WARNING');
      expect(recovered).toContain('recovered from parse error');
      expect(warnings.some(w => w.includes('parse error'))).toBe(true);
    });

    it('should preserve empty lines as-is', () => {
      const content = 'line1\n\nline2';
      const error = new Error('parse failed');
      const processor = (line) => `processed: ${line}`;

      const { recovered } = recovery.recoverFromParseError(content, error, processor);

      const lines = recovered.split('\n');
      // There should be an empty line preserved somewhere
      expect(lines.some(l => l.trim() === '')).toBe(true);
    });
  });
});
