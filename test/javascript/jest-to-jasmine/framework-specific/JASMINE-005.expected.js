describe('full example', () => {
  fdescribe('focused suite', () => {
    it('uses timers and spies', () => {
      jasmine.clock().install();
      const fn = jasmine.createSpy();
      setTimeout(fn, 500);
      jasmine.clock().tick(500);
      expect(fn).toHaveBeenCalledTimes(1);
      jasmine.clock().uninstall();
    });

    xit('skipped test', () => {
      expect(true).toBe(true);
    });
  });
});
