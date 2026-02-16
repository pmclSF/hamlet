import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('Timer mocking', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('advances timers by specific time', () => {
    const callback = vi.fn();
    setTimeout(callback, 2000);
    expect(callback).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1000);
    expect(callback).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('runs all pending timers', () => {
    const first = vi.fn();
    const second = vi.fn();
    setTimeout(first, 1000);
    setTimeout(second, 5000);
    vi.runAllTimers();
    expect(first).toHaveBeenCalled();
    expect(second).toHaveBeenCalled();
  });

  it('runs only pending timers', () => {
    const callback = vi.fn();
    setTimeout(callback, 1000);
    vi.runOnlyPendingTimers();
    expect(callback).toHaveBeenCalledTimes(1);
  });
});
