describe('test', () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it('works', () => {
    const fn = jest.fn();
    fn();
  });
});
