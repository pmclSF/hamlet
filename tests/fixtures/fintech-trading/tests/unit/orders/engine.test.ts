import { describe, it, expect } from 'vitest';
import { placeOrder, cancelOrder, fillOrder } from '../../../src/orders/engine';
describe('placeOrder', () => { it('should place', () => { expect(placeOrder('AAPL', 10).status).toBe('placed'); }); });
describe('cancelOrder', () => { it('should cancel', () => { expect(cancelOrder('ord_1').status).toBe('cancelled'); }); });
