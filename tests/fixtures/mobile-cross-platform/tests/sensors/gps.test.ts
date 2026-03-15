import { describe, it, expect } from 'vitest';
import { isValidLocation, distanceBetween } from '../../src/sensors/gps';

describe('GPS', () => {
  it('should validate location within bounds', () => {
    expect(isValidLocation({ latitude: 37.7749, longitude: -122.4194, accuracy: 10 })).toBe(true);
  });

  it('should reject invalid latitude', () => {
    expect(isValidLocation({ latitude: 91, longitude: 0, accuracy: 10 })).toBe(false);
  });

  it('should calculate distance between points', () => {
    const a = { latitude: 0, longitude: 0, accuracy: 10 };
    const b = { latitude: 1, longitude: 0, accuracy: 10 };
    const dist = distanceBetween(a, b);
    expect(dist).toBeGreaterThan(100000);
    expect(dist).toBeLessThan(120000);
  });

  it.skip('should read GPS from device sensor', () => {
    // TODO: requires device GPS hardware
  });
});
