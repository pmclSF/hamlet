import { describe, it, expect } from 'vitest';
import { detectPlatform, isNativePlatform, formatDeviceId } from '../../src/platform/shared';

describe('Shared Platform', () => {
  it('should detect platform', () => {
    expect(detectPlatform()).toBe('web');
  });

  it('should identify native platforms', () => {
    expect(isNativePlatform('ios')).toBe(true);
    expect(isNativePlatform('android')).toBe(true);
    expect(isNativePlatform('web')).toBe(false);
  });

  it('should format device id', () => {
    expect(formatDeviceId('ios', 'abc123')).toBe('ios:abc123');
  });
});
