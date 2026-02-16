describe('boolean edge cases with truthiness', () => {
  it('should treat zero as falsy', () => {
    expect(0).toBeFalsy();
    expect(0).not.toBeTruthy();
  });

  it('should treat empty string as falsy', () => {
    expect('').toBeFalsy();
    expect('non-empty').toBeTruthy();
  });

  it('should treat empty array as truthy', () => {
    expect([]).toBeTruthy();
    expect([]).not.toBeFalsy();
  });

  it('should treat null and undefined as falsy', () => {
    expect(null).toBeFalsy();
    expect(undefined).toBeFalsy();
  });

  it('should treat NaN as falsy', () => {
    expect(NaN).toBeFalsy();
    expect(NaN).not.toBeTruthy();
  });

  it('should treat non-zero numbers as truthy', () => {
    expect(1).toBeTruthy();
    expect(-1).toBeTruthy();
  });
});
