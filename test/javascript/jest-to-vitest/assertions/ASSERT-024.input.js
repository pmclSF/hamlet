describe('AsyncValidator', () => {
  it('should make exactly two assertions in the callback', () => {
    expect.assertions(2);
    return validateAsync({ name: 'test', value: 42 }).then((result) => {
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });
  });

  it('should assert inside a catch block', () => {
    expect.assertions(1);
    return validateAsync({ name: '' }).catch((error) => {
      expect(error.message).toMatch(/name is required/i);
    });
  });

  it('should make one assertion for hasAssertions', () => {
    expect.hasAssertions();
    const result = synchronousValidate('hello');
    expect(result).toBe(true);
  });
});
