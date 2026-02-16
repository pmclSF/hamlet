describe('timers', () => {
  beforeEach(() => {
    jasmine.clock().install();
  });
  afterEach(() => {
    jasmine.clock().uninstall();
  });
  it('advances time', () => {
    const fn = jasmine.createSpy();
    setTimeout(fn, 1000);
    jasmine.clock().tick(1000);
    expect(fn).toHaveBeenCalled();
  });
});
