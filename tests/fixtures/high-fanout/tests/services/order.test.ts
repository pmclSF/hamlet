import { describe, it, expect } from 'vitest';
import { connect, query, transaction } from '../../src/utils/database';
import { createOrder } from '../../src/services/order';

describe('Order Service', () => {
  it('should create order with transaction', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const order = transaction(() => createOrder(1, ['item-a', 'item-b']));
    expect(order.items).toHaveLength(2);
  });

  it('should query orders', () => {
    const result = query('SELECT * FROM orders WHERE user_id = ?', [1]);
    expect(result).toHaveLength(1);
  });
});
