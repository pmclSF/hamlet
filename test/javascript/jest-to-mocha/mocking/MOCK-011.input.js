describe('test', () => {
  it('timers', () => {
    jest.useFakeTimers();
    const fn = jest.fn();
    setTimeout(fn, 1000);
    jest.advanceTimersByTime(1000);
    expect(fn).toHaveBeenCalled();
    jest.useRealTimers();
  });
});
