describe('matchers', () => {
  it('uses anything', () => {
    const spy = jest.fn();
    spy(42);
    expect(spy).toHaveBeenCalledWith(expect.anything());
  });
});
