describe('clock', () => {
  it('mocks date', () => {
    jasmine.clock().install();
    jasmine.clock().mockDate(new Date(2020, 0, 1));
    expect(new Date().getFullYear()).toBe(2020);
    jasmine.clock().uninstall();
  });
});
