describe('HttpClient', () => {
  it('should reject with a network error', async () => {
    await expect(httpGet('http://unreachable.invalid')).rejects.toThrow('Network error');
  });

  it('should reject with a timeout error', async () => {
    await expect(httpGet('http://slow.example.com', { timeout: 1 }))
      .rejects.toThrow(/timeout/i);
  });

  it('should reject with the correct error type', async () => {
    await expect(httpGet('http://example.com/404'))
      .rejects.toBeInstanceOf(Error);
  });

  it('should reject when response is malformed', async () => {
    await expect(httpGetJson('http://example.com/not-json'))
      .rejects.toThrow('Invalid JSON response');
  });

  it('should reject with status code in error', async () => {
    try {
      await httpGet('http://example.com/500');
    } catch (error) {
      expect(error.statusCode).toBe(500);
      expect(error.message).toContain('Internal Server Error');
    }
  });
});
