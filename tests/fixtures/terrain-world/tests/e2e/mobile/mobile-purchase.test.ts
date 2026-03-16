import { describe, it, expect } from 'vitest';
import { initiatePurchase, completePurchase, validateReceipt } from '../../../src/mobile/purchase';
import { authenticate } from '../../../src/auth/login';
import { connectDB, seedTestData, cleanupDB } from '../../../src/shared-db';

describe('mobile purchase e2e', () => {
  it('should complete mobile purchase on iOS', () => {
    connectDB();
    seedTestData();
    authenticate('mobile@test.com', 'pass');
    const purchase = initiatePurchase('prod_1', 'ios');
    const receipt = validateReceipt('receipt_ios', 'ios');
    expect(receipt.valid).toBe(true);
    completePurchase(purchase.purchaseId);
    cleanupDB();
  });

  it('should complete mobile purchase on Android', () => {
    connectDB();
    seedTestData();
    authenticate('mobile@test.com', 'pass');
    const purchase = initiatePurchase('prod_1', 'android');
    const receipt = validateReceipt('receipt_android', 'android');
    expect(receipt.valid).toBe(true);
    completePurchase(purchase.purchaseId);
    cleanupDB();
  });
});
