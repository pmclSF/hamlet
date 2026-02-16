describe('spies', () => {
  it('gets all args', () => {
    const spy = jasmine.createSpy('fn');
    spy('a');
    spy('b');
    expect(spy.calls.allArgs()).toEqual([['a'], ['b']]);
  });
});
