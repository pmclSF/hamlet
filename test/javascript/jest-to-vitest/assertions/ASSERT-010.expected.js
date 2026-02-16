import { describe, it, expect } from 'vitest';

describe('AppSettings', () => {
  it('should set the default theme', () => {
    const settings = getDefaults();
    expect(settings).toHaveProperty('theme', 'light');
  });

  it('should have the correct max upload size', () => {
    const settings = getDefaults();
    expect(settings).toHaveProperty('upload.maxSizeMB', 25);
  });

  it('should have nested locale with country code', () => {
    const settings = getDefaults();
    expect(settings).toHaveProperty('locale.country', 'US');
  });

  it('should set the timeout to a number', () => {
    const settings = getDefaults();
    expect(settings).toHaveProperty('session.timeoutMs', 3600000);
  });

  it('should have feature flags as a boolean', () => {
    const settings = getDefaults();
    expect(settings).toHaveProperty('features.analytics', true);
  });
});
