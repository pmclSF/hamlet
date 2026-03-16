import { describe, it, expect } from 'vitest';
import { placeOrder, fillOrder } from '../../../src/orders/engine';
import { analyzerAction } from '../../../src/risk/analyzer';
import { trackerAction } from '../../../src/positions/tracker';
import { connectDB, seedAccount, cleanupDB } from '../../../src/shared/db';
describe('full trade e2e', () => {
  it('should complete trade', () => { connectDB(); seedAccount(); analyzerAction('check'); const o = placeOrder('AAPL', 5); fillOrder(o.orderId); trackerAction('update'); cleanupDB(); });
});
