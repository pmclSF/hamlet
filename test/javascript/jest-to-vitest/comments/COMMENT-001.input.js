describe('string formatter', () => {
  it('should capitalize the first letter', () => {
    // Arrange: set up input string
    const input = 'hello world';

    // Act: run the formatter
    const result = input.charAt(0).toUpperCase() + input.slice(1);

    // Assert: check the output
    expect(result).toBe('Hello world');
  });

  it('should handle empty strings gracefully', () => {
    const input = ''; // edge case: empty string
    const result = input.charAt(0).toUpperCase() + input.slice(1);
    expect(result).toBe(''); // should return empty string
  });
});
