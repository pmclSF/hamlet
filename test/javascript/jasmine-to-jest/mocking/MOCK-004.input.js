describe('spies', () => {
  it('returns value', () => {
    const spy = jasmine.createSpy('fn').and.returnValue(42);
    expect(spy()).toBe(42);
  });
});
