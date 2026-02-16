describe('spies', () => {
  it('gets first args', () => {
    const spy = jest.fn();
    spy('first');
    spy('second');
    expect(spy.mock.calls[0]).toEqual(['first']);
  });
});
