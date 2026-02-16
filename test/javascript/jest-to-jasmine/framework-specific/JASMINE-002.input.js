describe('timers', () => {
  it('mocks date', () => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date(2020, 0, 1));
    expect(new Date().getFullYear()).toBe(2020);
    jest.useRealTimers();
  });
});
