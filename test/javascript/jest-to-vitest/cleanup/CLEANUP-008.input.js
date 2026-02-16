describe('SharedResourceLifecycle', () => {
  let sharedPool;
  let connectionCount;

  beforeAll(() => {
    sharedPool = {
      connections: [],
      maxSize: 5,
      acquire() {
        const conn = { id: this.connections.length + 1, active: true };
        this.connections.push(conn);
        return conn;
      },
      releaseAll() {
        this.connections.forEach(c => { c.active = false; });
        this.connections = [];
      },
    };
    connectionCount = 0;
  });

  afterAll(() => {
    sharedPool.releaseAll();
    sharedPool = null;
  });

  it('should acquire a connection from the pool', () => {
    const conn = sharedPool.acquire();
    connectionCount++;
    expect(conn.active).toBe(true);
    expect(conn.id).toBe(connectionCount);
  });

  it('should track pool size', () => {
    sharedPool.acquire();
    connectionCount++;
    expect(sharedPool.connections.length).toBeGreaterThan(0);
    expect(sharedPool.connections.length).toBeLessThanOrEqual(sharedPool.maxSize);
  });

  it('should support multiple connections', () => {
    const conn1 = sharedPool.acquire();
    const conn2 = sharedPool.acquire();
    expect(conn1.id).not.toBe(conn2.id);
    expect(conn1.active).toBe(true);
    expect(conn2.active).toBe(true);
  });
});
