describe('spies', () => {
  it('gets args', () => {
    const spy = jasmine.createSpy('fn');
    spy('a', 'b');
    expect(spy.calls.argsFor(0)).toEqual(['a', 'b']);
  });
});
