describe('test', () => {
  it('fake timers', () => {
    jest.useFakeTimers();
    const fn = jest.fn();
    setTimeout(fn, 1000);
    jest.advanceTimersByTime(1000);
    expect(fn).toHaveBeenCalledTimes(1);
    jest.useRealTimers();
  });
});
