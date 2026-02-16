const mathUtils = {
  add: (a, b) => a + b,
  multiply: (a, b) => a * b,
};

describe('MathUtils', () => {
  it('spies on add method', () => {
    const spy = jest.spyOn(mathUtils, 'add').mockReturnValue(42);
    const result = mathUtils.add(1, 2);
    expect(result).toBe(42);
    expect(spy).toHaveBeenCalledWith(1, 2);
    spy.mockRestore();
  });

  it('spies on multiply without mocking', () => {
    const spy = jest.spyOn(mathUtils, 'multiply');
    mathUtils.multiply(3, 4);
    expect(spy).toHaveBeenCalledTimes(1);
    spy.mockRestore();
  });
});
