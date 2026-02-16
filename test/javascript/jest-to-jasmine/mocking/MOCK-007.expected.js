describe('mocks', () => {
  it('checks args', () => {
    const fn = jasmine.createSpy();
    fn('a', 'b');
    expect(fn.calls.argsFor(0)).toEqual(['a', 'b']);
  });
});
