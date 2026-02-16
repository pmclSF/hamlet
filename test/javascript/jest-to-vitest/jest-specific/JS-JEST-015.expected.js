import { describe, it, expect } from 'vitest';

expect.extend({
  toBeWithinRange(received, floor, ceiling) {
    const pass = received >= floor && received <= ceiling;
    return {
      pass,
      message: () => `expected ${received} to be within range ${floor} - ${ceiling}`,
    };
  },
});

describe('Custom matcher', () => {
  it('works for value in range', () => {
    expect(100).toBeWithinRange(90, 110);
  });

  it('fails for value out of range', () => {
    expect(200).not.toBeWithinRange(90, 110);
  });
});
