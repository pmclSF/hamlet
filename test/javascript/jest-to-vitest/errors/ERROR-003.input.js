describe('ConfigParser', () => {
  it('should not throw for a valid YAML config', () => {
    expect(() => parseConfig('database:\n  host: localhost')).not.toThrow();
  });

  it('should not throw for an empty config', () => {
    expect(() => parseConfig('')).not.toThrow();
  });

  it('should not throw when optional fields are missing', () => {
    expect(() => {
      const config = parseConfig('name: myapp');
      validateConfig(config, { strict: false });
    }).not.toThrow();
  });

  it('should not throw for a config with only comments', () => {
    expect(() => parseConfig('# This is a comment\n# Another comment')).not.toThrow();
  });
});
