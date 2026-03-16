import { describe, it, expect } from 'vitest';
import { placeOrder, fillOrder } from '../../../src/orders/engine';
import { analyzerAction } from '../../../src/risk/analyzer';
import { connectDB, seedAccount, cleanupDB } from '../../../src/shared/db';
describe('quick trade e2e', () => {
  it('should quick trade', () => { connectDB(); seedAccount(); analyzerAction('quick'); placeOrder('TSLA', 3); cleanupDB(); });
});
