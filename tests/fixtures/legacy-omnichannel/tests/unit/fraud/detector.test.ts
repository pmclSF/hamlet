import { describe, it, expect } from 'vitest';
import { analyzeRisk, checkVelocity } from '../../../src/fraud/detector';

describe('analyzeRisk', () => {
  it('should flag high amount', () => {
    expect(analyzeRisk('u1', 10000).flagged).toBe(true);
  });
  it('should not flag low amount', () => {
    expect(analyzeRisk('u1', 50).flagged).toBe(false);
  });
});

describe('checkVelocity', () => {
  it('should check velocity', () => {
    expect(checkVelocity('u1').flagged).toBe(false);
  });
});
