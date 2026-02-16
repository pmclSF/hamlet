describe('test', () => {
  it('calledOnce', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
