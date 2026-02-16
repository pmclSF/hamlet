import { describe, it, expect } from 'vitest';

expect.extend({
  toBeWithinRange(received, floor, ceiling) {
    const pass = received >= floor && received <= ceiling;
    if (pass) {
      return {
        message: () => `expected ${received} not to be within range ${floor} - ${ceiling}`,
        pass: true,
      };
    } else {
      return {
        message: () => `expected ${received} to be within range ${floor} - ${ceiling}`,
        pass: false,
      };
    }
  },
});

describe('CustomMatchers', () => {
  it('should validate that a number is within range', () => {
    expect(100).toBeWithinRange(90, 110);
  });

  it('should validate temperature readings', () => {
    const temp = readSensor();
    expect(temp).toBeWithinRange(-40, 85);
  });

  it('should support negation of custom matchers', () => {
    expect(200).not.toBeWithinRange(0, 100);
  });
});
