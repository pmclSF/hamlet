describe('spies', () => {
  it('gets most recent', () => {
    const spy = jest.fn();
    spy('first');
    spy('second');
    expect(spy.mock.lastCall).toEqual(['second']);
  });
});
