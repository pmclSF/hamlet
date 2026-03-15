import { describe, it, expect } from 'vitest';
import { vibrate, biometricAvailable, fcmToken } from '../../src/platform/android';

describe('Android Platform', () => {
  it('should accept valid vibration duration', () => {
    expect(vibrate(100)).toBe(true);
  });

  it('should reject zero duration', () => {
    expect(vibrate(0)).toBe(false);
  });

  it.skip('should detect biometric hardware', () => {
    // TODO: requires Android device with fingerprint sensor
    expect(biometricAvailable()).toBe(true);
  });

  it.skip('should generate valid FCM token', () => {
    // TODO: requires Firebase Cloud Messaging connection
    const token = fcmToken();
    expect(token).toMatch(/^fcm_/);
  });
});
