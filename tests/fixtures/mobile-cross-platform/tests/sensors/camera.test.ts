import { describe, it, expect } from 'vitest';
import { createCameraConfig, validateResolution } from '../../src/sensors/camera';

describe('Camera', () => {
  it('should create rear camera config with flash', () => {
    const config = createCameraConfig('rear');
    expect(config.flash).toBe(true);
    expect(config.facing).toBe('rear');
  });

  it('should create front camera config without flash', () => {
    const config = createCameraConfig('front');
    expect(config.flash).toBe(false);
  });

  it.skip('should validate hardware-supported resolutions', () => {
    // TODO: requires device camera hardware detection
    expect(validateResolution('4k')).toBe(true);
  });

  it('should validate resolution strings', () => {
    expect(validateResolution('1080p')).toBe(true);
    expect(validateResolution('480p')).toBe(false);
  });
});
