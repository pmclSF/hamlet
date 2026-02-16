jest.mock('./api');

describe('test', () => {
  it('combined', () => {
    const fn = jest.fn().mockReturnValue(42);
    expect(fn()).toBe(42);
  });
});
