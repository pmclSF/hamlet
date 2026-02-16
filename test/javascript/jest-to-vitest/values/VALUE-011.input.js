describe('NaN comparison', () => {
  it('should detect NaN with toBeNaN', () => {
    expect(NaN).toBeNaN();
  });

  it('should detect NaN from invalid operations', () => {
    const result = parseInt('not-a-number', 10);
    expect(result).toBeNaN();
  });

  it('should confirm valid numbers are not NaN', () => {
    expect(42).not.toBeNaN();
    expect(0).not.toBeNaN();
    expect(-1).not.toBeNaN();
  });

  it('should detect NaN from math operations', () => {
    const result = Math.sqrt(-1);
    expect(result).toBeNaN();
    expect(0 / 0).toBeNaN();
  });
});
