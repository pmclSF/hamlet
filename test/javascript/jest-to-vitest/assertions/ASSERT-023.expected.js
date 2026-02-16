import { describe, it, expect, vi } from 'vitest';

describe('RetryManager', () => {
  it('should retry the operation exactly 3 times', () => {
    const operation = vi.fn(() => { throw new Error('fail'); });
    const manager = new RetryManager({ maxRetries: 3 });
    try { manager.execute(operation); } catch (_e) { /* expected */ }
    expect(operation).toHaveBeenCalledTimes(3);
  });

  it('should call the success handler once on success', () => {
    const onSuccess = vi.fn();
    const operation = vi.fn(() => 'ok');
    const manager = new RetryManager({ maxRetries: 3, onSuccess });
    manager.execute(operation);
    expect(onSuccess).toHaveBeenCalledTimes(1);
    expect(operation).toHaveBeenCalledTimes(1);
  });

  it('should call the failure callback after all retries exhausted', () => {
    const onFailure = vi.fn();
    const operation = vi.fn(() => { throw new Error('fail'); });
    const manager = new RetryManager({ maxRetries: 2, onFailure });
    try { manager.execute(operation); } catch (_e) { /* expected */ }
    expect(onFailure).toHaveBeenCalledTimes(1);
  });
});
