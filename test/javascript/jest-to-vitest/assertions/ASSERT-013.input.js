describe('InputValidator', () => {
  it('should throw TypeError for non-string input', () => {
    expect(() => validate(123)).toThrow(TypeError);
  });

  it('should throw RangeError for out-of-range values', () => {
    expect(() => validateAge(-1)).toThrow(RangeError);
  });

  it('should throw SyntaxError for malformed JSON', () => {
    expect(() => parseConfig('not-json')).toThrow(SyntaxError);
  });

  it('should throw a generic Error for missing required fields', () => {
    expect(() => validateForm({})).toThrow(Error);
  });
});
