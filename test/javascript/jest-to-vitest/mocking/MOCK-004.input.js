describe('Custom mock implementation', () => {
  it('uses a mock implementation for transformation', () => {
    const transform = jest.fn().mockImplementation((x) => x * 2);
    expect(transform(5)).toBe(10);
    expect(transform(3)).toBe(6);
    expect(transform).toHaveBeenCalledTimes(2);
  });

  it('uses mockImplementationOnce for single call', () => {
    const greet = jest.fn()
      .mockImplementationOnce((name) => `Hello, ${name}!`)
      .mockImplementationOnce((name) => `Hi, ${name}!`);
    expect(greet('Alice')).toBe('Hello, Alice!');
    expect(greet('Bob')).toBe('Hi, Bob!');
  });
});
