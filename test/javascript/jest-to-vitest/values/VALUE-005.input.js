describe('special character handling', () => {
  it('should handle strings with brackets', () => {
    const selector = 'input[name="email"]';
    expect(selector).toContain('[name=');
    expect(selector).toBe('input[name="email"]');
  });

  it('should handle strings with escaped quotes', () => {
    const str = 'He said "hello" and she said \'goodbye\'';
    expect(str).toContain('"hello"');
    expect(str).toContain("'goodbye'");
  });

  it('should handle strings with colons and semicolons', () => {
    const css = 'color: red; font-size: 14px;';
    expect(css).toContain('color: red');
    expect(css).toContain('font-size: 14px');
  });

  it('should handle backslash characters', () => {
    const path = 'C:\\Users\\test\\file.txt';
    expect(path).toContain('\\Users\\');
    expect(path).toMatch(/\\file\.txt$/);
  });
});
