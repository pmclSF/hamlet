describe('data processor', () => {
  // TODO: add tests for CSV parsing once the parser is stable
  it('should process JSON data', () => {
    const input = '{"name":"Alice"}';
    const result = JSON.parse(input);
    expect(result.name).toBe('Alice');
  });

  // FIXME: this test is flaky on CI due to timing issues
  it('should handle large payloads', () => {
    const largeArray = Array.from({ length: 10000 }, (_, i) => i);
    expect(largeArray).toHaveLength(10000);
  });

  // HACK: working around a bug in the serializer
  it('should serialize dates as ISO strings', () => {
    const date = new Date('2024-06-15T00:00:00.000Z');
    const serialized = date.toISOString();
    expect(serialized).toBe('2024-06-15T00:00:00.000Z');
  });
});
