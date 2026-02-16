describe('clock', () => {
  beforeEach(() => {
    jasmine.clock().install();
  });
  afterEach(() => {
    jasmine.clock().uninstall();
  });
  it('ticks', () => {
    const fn = jasmine.createSpy('fn');
    setTimeout(fn, 1000);
    jasmine.clock().tick(1000);
    expect(fn).toHaveBeenCalled();
  });
});
