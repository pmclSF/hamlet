describe('spies', () => {
  it('multiple spies', () => {
    const spy1 = jest.fn();
    const spy2 = jest.fn();
    spy1('a');
    spy2('b');
    expect(spy1).toHaveBeenCalledWith('a');
    expect(spy2).toHaveBeenCalledWith('b');
  });
});
