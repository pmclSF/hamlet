import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Audit Service', () => {
  it('should log audit events', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('INSERT INTO audit_log (action, user_id) VALUES (?, ?)', ['login', 1]);
    expect(result).toBeDefined();
  });

  it('should query audit trail', () => {
    const result = query('SELECT * FROM audit_log WHERE timestamp > ?', ['2024-01-01']);
    expect(result).toHaveLength(1);
  });
});
