import { describe, it, expect } from 'vitest';
import { analyzeTransaction, checkVelocity, reportFraud } from '../../../src/fraud/detector';

describe('analyzeTransaction', () => {
  it('should analyze transaction', () => {
    const result = analyzeTransaction('tx_1', 500);
    expect(result).toBeTruthy();
  });

  it('should flag high-risk transactions', () => {
    const result = analyzeTransaction('tx_2', 15000);
    expect(result).toBeTruthy();
  });
});

describe('checkVelocity', () => {
  it('should check velocity', () => {
    expect(checkVelocity('user_1')).toBeTruthy();
  });
});

describe('reportFraud', () => {
  it('should report fraud', () => {
    expect(reportFraud('tx_1', 'suspicious')).toBeTruthy();
  });
});
