describe('Counter', () => {
  let count;
  let initialValue;

  beforeEach(() => {
    initialValue = 10;
    count = initialValue;
  });

  it('should start at the initial value', () => {
    expect(count).toBe(10);
  });

  it('should increment correctly', () => {
    count += 1;
    expect(count).toBe(initialValue + 1);
  });

  it('should decrement correctly', () => {
    count -= 3;
    expect(count).toBe(initialValue - 3);
  });

  it('should be isolated between tests', () => {
    // Even though previous tests modified count, beforeEach resets it
    expect(count).toBe(initialValue);
  });
});
