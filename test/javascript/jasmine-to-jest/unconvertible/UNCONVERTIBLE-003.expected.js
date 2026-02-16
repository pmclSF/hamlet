describe('mixed', () => {
  it('has both', () => {
    const spy = jest.fn();
    spy();
    expect(spy).toHaveBeenCalled();
  });
});
