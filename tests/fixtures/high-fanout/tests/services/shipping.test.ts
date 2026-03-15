import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Shipping Service', () => {
  it('should query shipment status', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT * FROM shipments WHERE order_id = ?', [1]);
    expect(result).toHaveLength(1);
  });

  it('should track delivery', () => {
    const result = query('SELECT status FROM shipments WHERE tracking = ?', ['TRK001']);
    expect(result).toBeDefined();
  });
});
