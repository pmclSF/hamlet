describe('timers', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });
  afterEach(() => {
    jest.useRealTimers();
  });
  it('advances time', () => {
    const fn = jest.fn();
    setTimeout(fn, 1000);
    jest.advanceTimersByTime(1000);
    expect(fn).toHaveBeenCalled();
  });
});
