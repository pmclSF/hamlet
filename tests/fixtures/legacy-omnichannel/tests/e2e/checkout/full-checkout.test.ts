import { describe, it, expect } from 'vitest';
import { initiateCheckout, completeCheckout } from '../../../src/checkout/checkout';
import { analyzeRisk } from '../../../src/fraud/detector';
import { getAdminStats } from '../../../src/admin/dashboard';
import { connectDB, seedTestData, createTestUser, createTestOrder, cleanupDB } from '../../../src/shared/db-helper';

describe('full checkout e2e', () => {
  it('should complete end-to-end checkout', () => {
    connectDB(); seedTestData(); createTestUser(); createTestOrder();
    const risk = analyzeRisk('usr_test', 99);
    expect(risk.flagged).toBe(false);
    const co = initiateCheckout('usr_test', 'cart_1');
    const done = completeCheckout(co.checkoutId, 'pay_valid');
    expect(done.status).toBe('completed');
    const stats = getAdminStats();
    expect(stats.totalOrders).toBeGreaterThan(0);
    cleanupDB();
  });
});
