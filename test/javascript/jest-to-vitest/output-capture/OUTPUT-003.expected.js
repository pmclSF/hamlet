import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('NoisyModule', () => {
  let logSpy;
  let errorSpy;

  beforeEach(() => {
    logSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    logSpy.mockRestore();
    errorSpy.mockRestore();
  });

  it('should perform operation without polluting test output', () => {
    const noisyProcess = () => {
      console.log('Starting process...');
      console.log('Processing item 1 of 100...');
      console.error('Non-critical warning: cache miss');
      console.log('Done.');
      return { success: true, count: 100 };
    };

    const result = noisyProcess();
    expect(result.success).toBe(true);
    expect(result.count).toBe(100);
  });

  it('should suppress verbose initialization logs', () => {
    const init = () => {
      console.log('Loading config...');
      console.log('Connecting to services...');
      return { initialized: true };
    };

    const state = init();
    expect(state.initialized).toBe(true);
  });
});
