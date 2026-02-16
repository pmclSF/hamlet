describe('Timer mocking', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('advances timers by specific time', () => {
    const callback = jest.fn();
    setTimeout(callback, 2000);
    expect(callback).not.toHaveBeenCalled();
    jest.advanceTimersByTime(1000);
    expect(callback).not.toHaveBeenCalled();
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('runs all pending timers', () => {
    const first = jest.fn();
    const second = jest.fn();
    setTimeout(first, 1000);
    setTimeout(second, 5000);
    jest.runAllTimers();
    expect(first).toHaveBeenCalled();
    expect(second).toHaveBeenCalled();
  });

  it('runs only pending timers', () => {
    const callback = jest.fn();
    setTimeout(callback, 1000);
    jest.runOnlyPendingTimers();
    expect(callback).toHaveBeenCalledTimes(1);
  });
});
