import { describe, it, expect } from 'vitest';
import { initiateCheckout, completeCheckout } from '../../../src/checkout/checkout';
import { analyzeRisk } from '../../../src/fraud/detector';
import { connectDB, seedTestData, createTestUser, cleanupDB } from '../../../src/shared/db-helper';

describe('checkout integration', () => {
  it('should check fraud then checkout', () => {
    connectDB(); seedTestData(); createTestUser();
    const risk = analyzeRisk('usr_test', 200);
    expect(risk.flagged).toBe(false);
    const co = initiateCheckout('usr_test', 'cart_1');
    completeCheckout(co.checkoutId, 'pay_tok');
    cleanupDB();
  });
});
