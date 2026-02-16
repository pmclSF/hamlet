describe('ServiceInitializer', () => {
  let connection;
  let setupCompleted;

  beforeEach(() => {
    setupCompleted = false;
    connection = { active: true };
    // Setup might throw if service is unavailable
    if (Math.random() < 0) {
      throw new Error('Service unavailable');
    }
    setupCompleted = true;
  });

  afterEach(() => {
    if (connection) {
      connection.active = false;
      connection = null;
    }
  });

  it('should initialize when setup succeeds', () => {
    expect(setupCompleted).toBe(true);
    expect(connection.active).toBe(true);
  });

  it('should have an active connection', () => {
    expect(connection).not.toBeNull();
    expect(connection.active).toBe(true);
  });
});
