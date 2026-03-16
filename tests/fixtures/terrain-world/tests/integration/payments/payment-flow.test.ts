import { describe, it, expect } from 'vitest';
import { createCharge, captureCharge } from '../../../src/payments/charge';
import { connectDB, seedTestData, createTestUser, cleanupDB } from '../../../src/shared-db';

describe('payment integration', () => {
  it('should create and capture charge', () => {
    connectDB();
    seedTestData();
    createTestUser();
    const charge = createCharge(1000, 'usd');
    const captured = captureCharge(charge.chargeId);
    expect(captured.status).toBe('captured');
    cleanupDB();
  });
});
