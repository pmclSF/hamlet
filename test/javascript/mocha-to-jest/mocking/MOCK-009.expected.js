describe('test', () => {
  it('notCalled', () => {
    const fn = jest.fn();
    expect(fn).not.toHaveBeenCalled();
  });
});
