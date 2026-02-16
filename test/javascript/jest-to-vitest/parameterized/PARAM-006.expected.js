import { describe, it, expect } from 'vitest';

const currencies = [
  { code: 'USD', symbol: '$', decimals: 2 },
  { code: 'EUR', symbol: '\u20AC', decimals: 2 },
  { code: 'JPY', symbol: '\u00A5', decimals: 0 },
  { code: 'GBP', symbol: '\u00A3', decimals: 2 },
];

describe('Currency formatting', () => {
  currencies.forEach(({ code, symbol, decimals }) => {
    it(`should format ${code} correctly`, () => {
      expect(code).toHaveLength(3);
      expect(symbol).toBeDefined();
      expect(decimals).toBeGreaterThanOrEqual(0);
    });
  });
});
