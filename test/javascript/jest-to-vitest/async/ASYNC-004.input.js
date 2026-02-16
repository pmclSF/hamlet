describe('Timer utilities', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('executes callback after delay', () => {
    const callback = jest.fn();
    delayedCall(callback, 1000);
    expect(callback).not.toHaveBeenCalled();
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('runs all pending timers', () => {
    const callback = jest.fn();
    delayedCall(callback, 5000);
    jest.runAllTimers();
    expect(callback).toHaveBeenCalled();
  });
});
