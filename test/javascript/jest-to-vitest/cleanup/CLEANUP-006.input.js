describe('MultiResourceCleanup', () => {
  it('should clean up multiple resources with try/finally', () => {
    const db = { connected: true, queries: [] };
    const cache = { entries: new Map(), active: true };
    const logger = { messages: [] };

    try {
      db.queries.push('SELECT * FROM users');
      cache.entries.set('user:1', { name: 'Alice' });
      logger.messages.push('Operation started');

      expect(db.queries).toHaveLength(1);
      expect(cache.entries.size).toBe(1);
      expect(logger.messages).toHaveLength(1);
    } finally {
      db.connected = false;
      db.queries.length = 0;
      cache.entries.clear();
      cache.active = false;
      logger.messages.length = 0;
    }

    expect(db.connected).toBe(false);
    expect(cache.active).toBe(false);
    expect(logger.messages).toHaveLength(0);
  });

  it('should handle nested try/finally blocks', () => {
    const outer = { open: true };
    const inner = { open: true };

    try {
      expect(outer.open).toBe(true);
      try {
        expect(inner.open).toBe(true);
        inner.data = 'processed';
      } finally {
        inner.open = false;
      }
      expect(inner.open).toBe(false);
    } finally {
      outer.open = false;
    }

    expect(outer.open).toBe(false);
  });
});
