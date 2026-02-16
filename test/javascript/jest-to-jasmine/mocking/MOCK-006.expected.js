describe('mocks', () => {
  it('checks call count', () => {
    const fn = jasmine.createSpy();
    fn();
    fn();
    expect(fn.calls.count()).toBe(2);
  });
});
