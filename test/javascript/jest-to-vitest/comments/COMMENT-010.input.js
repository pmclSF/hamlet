describe('URL parser', () => {
  // See: https://github.com/org/repo/issues/123
  it('should parse query parameters correctly', () => {
    const url = 'https://example.com?foo=bar&baz=qux';
    const params = new URLSearchParams(url.split('?')[1]);
    expect(params.get('foo')).toBe('bar');
    expect(params.get('baz')).toBe('qux');
  });

  // Related: https://github.com/org/repo/issues/456
  // This was a regression introduced in v2.3.0
  it('should handle URLs without query strings', () => {
    const url = 'https://example.com/path';
    const queryPart = url.split('?')[1];
    expect(queryPart).toBeUndefined();
  });

  // Ref: JIRA-789 - encoding edge case
  it('should decode encoded characters', () => {
    const encoded = 'hello%20world';
    expect(decodeURIComponent(encoded)).toBe('hello world');
  });
});
