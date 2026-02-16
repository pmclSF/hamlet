describe('combined', () => {
  it('uses clock and spy together', () => {
    jest.useFakeTimers();
    const spy = jest.fn();
    setTimeout(spy, 500);
    jest.advanceTimersByTime(500);
    expect(spy).toHaveBeenCalledTimes(1);
    jest.useRealTimers();
  });
});
