import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Search Service', () => {
  it('should search products', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT * FROM products WHERE name LIKE ?', ['%widget%']);
    expect(result).toHaveLength(1);
  });

  it('should search with filters', () => {
    const result = query('SELECT * FROM products WHERE category = ? AND price < ?', ['electronics', 1000]);
    expect(result).toBeDefined();
  });
});
