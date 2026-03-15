import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Admin Service', () => {
  it('should list admin users', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('SELECT * FROM users WHERE role = ?', ['admin']);
    expect(result).toHaveLength(1);
  });

  it('should audit admin actions', () => {
    const result = query('SELECT * FROM audit_log WHERE actor_role = ?', ['admin']);
    expect(result).toBeDefined();
  });
});
