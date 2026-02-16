describe('mocks', () => {
  it('returns value', () => {
    const fn = jasmine.createSpy().and.returnValue(42);
    expect(fn()).toBe(42);
  });
});
