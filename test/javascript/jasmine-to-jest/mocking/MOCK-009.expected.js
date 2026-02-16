describe('spies', () => {
  it('gets args', () => {
    const spy = jest.fn();
    spy('a', 'b');
    expect(spy.mock.calls[0]).toEqual(['a', 'b']);
  });
});
