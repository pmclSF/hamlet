describe('spies', () => {
  it('creates a spy', () => {
    const spy = jest.fn();
    spy();
    expect(spy).toHaveBeenCalled();
  });
});
