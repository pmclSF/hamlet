import { describe, it, expect } from 'vitest';
import { connect, query, transaction } from '../../src/utils/database';

describe('Inventory Service', () => {
  it('should check stock levels', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT quantity FROM inventory WHERE sku = ?', ['WIDGET-1']);
    expect(result).toHaveLength(1);
  });

  it('should reserve stock in transaction', () => {
    const result = transaction(() => query('UPDATE inventory SET reserved = reserved + 1'));
    expect(result).toBeDefined();
  });
});
