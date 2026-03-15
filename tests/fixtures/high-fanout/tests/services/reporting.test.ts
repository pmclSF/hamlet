import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Reporting Service', () => {
  it('should generate monthly report data', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT * FROM reports WHERE month = ?', ['2024-01']);
    expect(result).toHaveLength(1);
  });

  it('should calculate revenue summary', () => {
    const result = query('SELECT SUM(revenue) FROM monthly_summary');
    expect(result).toBeDefined();
  });
});
