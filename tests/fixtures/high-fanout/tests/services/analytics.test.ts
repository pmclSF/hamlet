import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Analytics Service', () => {
  it('should query page views', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT COUNT(*) FROM page_views WHERE date > ?', ['2024-01-01']);
    expect(result).toBeDefined();
  });

  it('should aggregate conversion metrics', () => {
    const result = query('SELECT conversion_rate FROM metrics');
    expect(result).toHaveLength(1);
  });
});
