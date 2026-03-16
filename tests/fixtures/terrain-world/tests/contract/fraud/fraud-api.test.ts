import { describe, it, expect } from 'vitest';
import { analyzeTransaction } from '../../../src/fraud/detector';

describe('fraud API contract', () => {
  it('analyzeTransaction returns required fields', () => {
    const result = analyzeTransaction('tx_1', 500);
    expect(result).toHaveProperty('transactionId');
    expect(result).toHaveProperty('riskScore');
    expect(result).toHaveProperty('flagged');
    expect(typeof result.riskScore).toBe('number');
    expect(typeof result.flagged).toBe('boolean');
  });
});
