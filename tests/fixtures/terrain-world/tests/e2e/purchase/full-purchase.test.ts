import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { createSession } from '../../../src/auth/session';
import { createCharge, captureCharge } from '../../../src/payments/charge';
import { analyzeTransaction } from '../../../src/fraud/detector';
import { sendEmail } from '../../../src/notifications/email';
import { connectDB, seedTestData, createTestUser, createTestPayment, cleanupDB } from '../../../src/shared-db';

describe('full purchase e2e', () => {
  it('should complete end-to-end purchase flow', () => {
    connectDB();
    seedTestData();
    createTestUser();
    const auth = authenticate('buyer@test.com', 'pass');
    const session = createSession(auth.token);
    const charge = createCharge(5000, 'usd');
    const fraud = analyzeTransaction(charge.chargeId, 5000);
    expect(fraud.flagged).toBe(false);
    const captured = captureCharge(charge.chargeId);
    expect(captured.status).toBe('captured');
    sendEmail('buyer@test.com', 'Receipt', 'Your purchase is confirmed');
    cleanupDB();
  });
});
