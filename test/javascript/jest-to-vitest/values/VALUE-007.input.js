describe('null and undefined comparisons', () => {
  it('should assert null values', () => {
    const result = null;
    expect(result).toBeNull();
    expect(result).not.toBeUndefined();
  });

  it('should assert undefined values', () => {
    let value;
    expect(value).toBeUndefined();
    expect(value).not.toBeNull();
  });

  it('should assert defined values', () => {
    const obj = { key: 'value' };
    expect(obj.key).toBeDefined();
    expect(obj.missing).toBeUndefined();
  });

  it('should use not.toBeNull for non-null checks', () => {
    const result = 0;
    expect(result).not.toBeNull();
    expect(result).toBeDefined();
  });
});
