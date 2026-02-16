jest.mock('./api');

describe('test', () => {
  it('full', () => {
    const fn = jest.fn().mockReturnValue(42);
    const result = fn();
    expect(result).toBe(42);
    expect(fn).toHaveBeenCalledTimes(1);
  });
});
