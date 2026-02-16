import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('Timer functions', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('calls the callback after 1 second', () => {
    const callback = vi.fn();
    setTimeout(callback, 1000);
    expect(callback).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('runs all pending timers', () => {
    const callback = vi.fn();
    setTimeout(callback, 5000);
    vi.runAllTimers();
    expect(callback).toHaveBeenCalled();
  });
});
