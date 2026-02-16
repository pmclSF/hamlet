import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('DeprecationWarner', () => {
  let warnSpy;

  beforeEach(() => {
    warnSpy = vi.spyOn(console, 'warn').mockImplementation();
  });

  afterEach(() => {
    warnSpy.mockRestore();
  });

  it('should warn about deprecated methods', () => {
    const api = {
      oldMethod() {
        console.warn('DEPRECATED: oldMethod() will be removed in v3.0');
        return 'result';
      },
    };

    const result = api.oldMethod();

    expect(result).toBe('result');
    expect(warnSpy).toHaveBeenCalledWith(
      'DEPRECATED: oldMethod() will be removed in v3.0'
    );
  });

  it('should warn about invalid configuration', () => {
    const config = {
      validate(opts) {
        if (opts.timeout < 0) {
          console.warn('Warning: negative timeout, using default');
        }
        return { ...opts, timeout: Math.max(opts.timeout, 1000) };
      },
    };

    const result = config.validate({ timeout: -1 });

    expect(warnSpy).toHaveBeenCalledTimes(1);
    expect(result.timeout).toBe(1000);
  });
});
