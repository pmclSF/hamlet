import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';
import { createUser } from '../../src/services/user';

describe('User Service', () => {
  it('should create user with database connection', () => {
    const db = connect({ host: 'localhost', port: 5432, database: 'test' });
    expect(db.connected).toBe(true);
    const user = createUser('Alice', 'alice@example.com');
    expect(user.name).toBe('Alice');
  });

  it('should query users', () => {
    const result = query('SELECT * FROM users');
    expect(result).toHaveLength(1);
  });
});
