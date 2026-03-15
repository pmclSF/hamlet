import { describe, it, expect } from 'vitest';
import { hapticFeedback, faceIdAvailable, pushNotificationToken } from '../../src/platform/ios';

describe('iOS Platform', () => {
  it('should provide haptic feedback for light intensity', () => {
    expect(hapticFeedback('light')).toBe(true);
  });

  it('should provide haptic feedback for medium intensity', () => {
    expect(hapticFeedback('medium')).toBe(true);
  });

  it.skip('should reject heavy haptic on older devices', () => {
    // TODO: requires iOS 14+ device
    expect(hapticFeedback('heavy')).toBe(false);
  });

  it.skip('should check Face ID availability on device', () => {
    // TODO: requires physical iOS device with Face ID
    expect(faceIdAvailable()).toBe(true);
  });

  it.skip('should generate valid APNS token', () => {
    // TODO: requires Apple Push Notification service connection
    const token = pushNotificationToken();
    expect(token).toMatch(/^apns_/);
  });
});
