describe('test', () => {
  it('chai-sinon calledWith', () => {
    const fn = jest.fn();
    fn('x');
    expect(fn).toHaveBeenCalledWith('x');
  });
});
