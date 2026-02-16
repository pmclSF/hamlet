describe('combined', () => {
  it('uses clock and spy together', () => {
    jasmine.clock().install();
    const spy = jasmine.createSpy('fn');
    setTimeout(spy, 500);
    jasmine.clock().tick(500);
    expect(spy).toHaveBeenCalledTimes(1);
    jasmine.clock().uninstall();
  });
});
