describe('spies', () => {
  it('throws error', () => {
    const spy = jasmine.createSpy('fn').and.throwError('oops');
    expect(() => spy()).toThrow();
  });
});
