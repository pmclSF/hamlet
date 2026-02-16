describe('spies', () => {
  it('checks call count', () => {
    const spy = jest.fn();
    spy();
    spy();
    expect(spy.mock.calls.length).toBe(2);
  });
});
