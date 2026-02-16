describe('mocks', () => {
  it('implements', () => {
    const fn = jasmine.createSpy().and.callFake(x => x * 2);
    expect(fn(5)).toBe(10);
  });
});
