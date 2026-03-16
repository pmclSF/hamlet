import { describe, it, expect } from 'vitest';
import { initiateCheckout, completeCheckout } from '../../../src/checkout/checkout';
import { analyzeRisk } from '../../../src/fraud/detector';
import { connectDB, seedTestData, createTestUser, cleanupDB } from '../../../src/shared/db-helper';

describe('express checkout e2e', () => {
  it('should do fast checkout', () => {
    connectDB(); seedTestData(); createTestUser();
    analyzeRisk('usr_test', 50);
    const co = initiateCheckout('usr_test', 'cart_1');
    completeCheckout(co.checkoutId, 'pay_tok');
    cleanupDB();
  });
});
