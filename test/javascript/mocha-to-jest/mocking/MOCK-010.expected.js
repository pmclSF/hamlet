describe('test', () => {
  it('timers', () => {
    jest.useFakeTimers();
    setTimeout(() => {}, 1000);
    jest.advanceTimersByTime(1000);
    jest.useRealTimers();
  });
});
