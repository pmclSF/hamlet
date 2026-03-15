import { describe, it, expect } from 'vitest';
import { connect, query } from '../../src/utils/database';

describe('Notification Service', () => {
  it('should queue notification', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = query('INSERT INTO notifications (user_id, message) VALUES (?, ?)', [1, 'Hello']);
    expect(result).toBeDefined();
  });

  it('should fetch unread notifications', () => {
    const result = query('SELECT * FROM notifications WHERE read = false');
    expect(result).toHaveLength(1);
  });
});
