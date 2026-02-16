describe('spies', () => {
  it('gets all args', () => {
    const spy = jest.fn();
    spy('a');
    spy('b');
    expect(spy.mock.calls).toEqual([['a'], ['b']]);
  });
});
