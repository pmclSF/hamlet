describe('full example', () => {
  describe.only('focused suite', () => {
    it('uses timers and spies', () => {
      jest.useFakeTimers();
      const fn = jest.fn();
      setTimeout(fn, 500);
      jest.advanceTimersByTime(500);
      expect(fn).toHaveBeenCalledTimes(1);
      jest.useRealTimers();
    });

    it.skip('skipped test', () => {
      expect(true).toBe(true);
    });
  });
});
