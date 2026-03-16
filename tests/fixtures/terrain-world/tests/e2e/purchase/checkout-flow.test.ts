import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { createSession } from '../../../src/auth/session';
import { createCharge, captureCharge } from '../../../src/payments/charge';
import { analyzeTransaction } from '../../../src/fraud/detector';
import { connectDB, seedTestData, createTestUser, cleanupDB } from '../../../src/shared-db';

describe('checkout e2e', () => {
  it('should complete checkout flow', () => {
    connectDB();
    seedTestData();
    createTestUser();
    const auth = authenticate('shopper@test.com', 'pass');
    createSession(auth.token);
    const charge = createCharge(3000, 'usd');
    const fraud = analyzeTransaction(charge.chargeId, 3000);
    expect(fraud.flagged).toBe(false);
    captureCharge(charge.chargeId);
    cleanupDB();
  });
});
