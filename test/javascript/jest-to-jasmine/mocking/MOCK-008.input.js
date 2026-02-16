describe('mocks', () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('works', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
