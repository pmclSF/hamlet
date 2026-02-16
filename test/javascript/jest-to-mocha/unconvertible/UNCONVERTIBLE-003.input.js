jest.mock('./utils');

describe('test', () => {
  it('works', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
