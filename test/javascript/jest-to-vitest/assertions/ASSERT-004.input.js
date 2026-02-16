describe('FeatureFlags', () => {
  it('should return truthy for enabled features', () => {
    const flag = getFeatureFlag('darkMode');
    expect(flag).toBeTruthy();
  });

  it('should return falsy for disabled features', () => {
    const flag = getFeatureFlag('betaFeature');
    expect(flag).toBeFalsy();
  });

  it('should return null for unknown features', () => {
    const flag = getFeatureFlag('nonexistent');
    expect(flag).toBeNull();
  });

  it('should return undefined when no key is provided', () => {
    const flag = getFeatureFlag();
    expect(flag).toBeUndefined();
  });

  it('should have a defined config object', () => {
    const config = getFeatureConfig('darkMode');
    expect(config).toBeDefined();
  });
});
