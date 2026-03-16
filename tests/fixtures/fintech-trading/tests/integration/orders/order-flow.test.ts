import { describe, it, expect } from 'vitest';
import { placeOrder, fillOrder } from '../../../src/orders/engine';
import { connectDB, seedAccount, seedOrder, cleanupDB } from '../../../src/shared/db';
describe('order flow', () => {
  it('should place and fill', () => { connectDB(); seedAccount(); const o = placeOrder('AAPL', 10); fillOrder(o.orderId); cleanupDB(); });
});
