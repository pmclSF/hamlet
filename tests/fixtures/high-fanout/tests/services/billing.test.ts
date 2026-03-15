import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Billing Service', () => {
  it('should query billing records', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT * FROM billing');
    expect(result).toHaveLength(1);
  });

  it('should calculate invoice totals', () => {
    const result = query('SELECT SUM(amount) FROM invoices WHERE user_id = ?', [1]);
    expect(result).toBeDefined();
  });
});
