import { describe, it, expect } from 'vitest';
import { analyzeRisk, checkVelocity, reportFraud } from '../../../src/fraud/detector';
import { connectDB, getUser, cleanupDB } from '../../../src/shared/db-helper';

describe('fraud screening integration', () => {
  it('should analyze and check velocity', () => {
    connectDB();
    const user = getUser('usr_1');
    const risk = analyzeRisk(user.id, 500);
    const vel = checkVelocity(user.id);
    expect(risk.flagged).toBe(false);
    expect(vel.flagged).toBe(false);
    cleanupDB();
  });
});
