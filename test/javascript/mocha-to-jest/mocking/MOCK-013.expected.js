describe('test', () => {
  it('chai-sinon', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
