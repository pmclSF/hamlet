/**
 * Shorthands ↔ ConverterFactory sync test.
 *
 * Asserts that every direction from ConverterFactory.getSupportedConversions()
 * has corresponding shorthand aliases, and vice versa.
 */
import { ConverterFactory } from '../../src/core/ConverterFactory.js';
import {
  SHORTHANDS,
  CONVERSION_CATEGORIES,
} from '../../src/cli/shorthands.js';

describe('Shorthands sync with ConverterFactory', () => {
  const factoryDirections = ConverterFactory.getSupportedConversions();

  it('should have a shorthand for every supported conversion', () => {
    const shorthandDirections = new Set(
      Object.values(SHORTHANDS).map((s) => `${s.from}-${s.to}`)
    );

    for (const direction of factoryDirections) {
      expect(shorthandDirections).toContain(direction);
    }
  });

  it('should not have shorthands for unsupported conversions', () => {
    const factorySet = new Set(factoryDirections);

    for (const { from, to } of Object.values(SHORTHANDS)) {
      expect(factorySet).toContain(`${from}-${to}`);
    }
  });

  it('CONVERSION_CATEGORIES should cover all directions', () => {
    const categorized = new Set();
    for (const cat of CONVERSION_CATEGORIES) {
      for (const dir of cat.directions) {
        categorized.add(`${dir.from}-${dir.to}`);
      }
    }

    for (const direction of factoryDirections) {
      expect(categorized).toContain(direction);
    }
  });

  it('should have exactly 25 unique directions', () => {
    expect(factoryDirections).toHaveLength(25);

    const shorthandDirections = [
      ...new Set(
        Object.values(SHORTHANDS).map((s) => `${s.from}-${s.to}`)
      ),
    ];
    expect(shorthandDirections).toHaveLength(25);
  });
});
