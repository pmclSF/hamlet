type Config = { timeout: number; retries: number };

describe('Config', () => {
  it('should use satisfies', () => {
    const config = { timeout: 5000, retries: 3 } satisfies Config;
    expect(config.timeout).toBe(5000);
  });

  it('should enforce type shape', () => {
    const config = { timeout: 1000, retries: 0 } satisfies Config;
    expect(config.retries).toBe(0);
  });
});
