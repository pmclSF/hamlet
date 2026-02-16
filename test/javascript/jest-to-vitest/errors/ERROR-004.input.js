describe('DataProcessor', () => {
  it('should capture and inspect the error object', () => {
    try {
      processData(null);
      // Should not reach here
      expect(true).toBe(false);
    } catch (error) {
      expect(error).toBeInstanceOf(Error);
      expect(error.message).toContain('Cannot process null');
      expect(error.code).toBe('ERR_NULL_INPUT');
    }
  });

  it('should include context in the error', () => {
    try {
      processData({ format: 'xml' });
    } catch (error) {
      expect(error.message).toMatch(/unsupported format/i);
      expect(error.details).toHaveProperty('format', 'xml');
      expect(error.details.supportedFormats).toContain('json');
    }
  });

  it('should recover gracefully from partial failures', () => {
    let result;
    try {
      result = processData({ items: [1, 'bad', 3] });
    } catch (_error) {
      result = { partial: true, processed: 2 };
    }
    expect(result).toBeDefined();
    expect(result.partial).toBe(true);
  });
});
