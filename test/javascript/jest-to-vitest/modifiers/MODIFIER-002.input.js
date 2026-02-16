describe.skip('LegacyModule', () => {
  it('should parse v1 format', () => {
    const data = { format: 'v1', payload: 'abc' };
    expect(data.format).toBe('v1');
  });

  it('should transform v1 to v2', () => {
    const v1 = { format: 'v1', payload: 'abc' };
    const v2 = { format: 'v2', payload: v1.payload.toUpperCase() };
    expect(v2.format).toBe('v2');
    expect(v2.payload).toBe('ABC');
  });
});

xdescribe('DeprecatedHelpers', () => {
  it('should format dates in legacy format', () => {
    const date = new Date('2024-01-15');
    const formatted = date.toISOString().split('T')[0];
    expect(formatted).toBe('2024-01-15');
  });

  it('should truncate strings to 10 characters', () => {
    const long = 'This is a very long string';
    const truncated = long.substring(0, 10);
    expect(truncated).toBe('This is a ');
    expect(truncated.length).toBe(10);
  });
});
