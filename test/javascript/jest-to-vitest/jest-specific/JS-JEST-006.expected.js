import { describe, it, expect, vi } from 'vitest';

vi.mock('./database');

const { getConnection } = require('./database');

describe('Database layer', () => {
  it('uses the auto-mock from __mocks__ directory', () => {
    const conn = getConnection();
    expect(conn).toBeDefined();
    expect(conn.query).toBeDefined();
  });

  it('returns mocked query results', async () => {
    const conn = getConnection();
    const results = await conn.query('SELECT * FROM users');
    expect(results).toEqual([]);
  });
});
