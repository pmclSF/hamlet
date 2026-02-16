describe('test', () => {
  it('not called', () => {
    const fn = jest.fn();
    expect(fn).not.toHaveBeenCalled();
  });
});
