describe('Timer functions', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('calls the callback after 1 second', () => {
    const callback = jest.fn();
    setTimeout(callback, 1000);
    expect(callback).not.toHaveBeenCalled();
    jest.advanceTimersByTime(1000);
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('runs all pending timers', () => {
    const callback = jest.fn();
    setTimeout(callback, 5000);
    jest.runAllTimers();
    expect(callback).toHaveBeenCalled();
  });
});
