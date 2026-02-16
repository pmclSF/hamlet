import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('Timer utilities', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('executes callback after delay', () => {
    const callback = vi.fn();
    delayedCall(callback, 1000);
    expect(callback).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('runs all pending timers', () => {
    const callback = vi.fn();
    delayedCall(callback, 5000);
    vi.runAllTimers();
    expect(callback).toHaveBeenCalled();
  });
});
