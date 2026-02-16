describe('empty string assertions', () => {
  it('should assert empty string with toBe', () => {
    const result = '';
    expect(result).toBe('');
  });

  it('should assert empty string length', () => {
    const str = '';
    expect(str).toHaveLength(0);
  });

  it('should differentiate empty from whitespace', () => {
    const empty = '';
    const whitespace = '  ';
    expect(empty).toBe('');
    expect(whitespace).not.toBe('');
    expect(whitespace.trim()).toBe('');
  });

  it('should handle empty string in array', () => {
    const arr = ['hello', '', 'world'];
    expect(arr[1]).toBe('');
    expect(arr.filter(s => s !== '')).toHaveLength(2);
  });
});
