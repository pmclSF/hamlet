describe('ErrorFactory', () => {
  it('should create a TypeError for invalid input', () => {
    const error = createError('type');
    expect(error).toBeInstanceOf(TypeError);
  });

  it('should create a RangeError for out-of-bounds values', () => {
    const error = createError('range');
    expect(error).toBeInstanceOf(RangeError);
  });

  it('should default to a generic Error', () => {
    const error = createError('generic');
    expect(error).toBeInstanceOf(Error);
  });

  it('should return Date instances from the parser', () => {
    const result = parseDate('2024-01-15');
    expect(result).toBeInstanceOf(Date);
  });
});
