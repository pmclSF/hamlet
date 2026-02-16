describe('DatabaseSetup', () => {
  let connection;

  beforeAll(async () => {
    try {
      connection = await connectToDatabase({ host: 'localhost', port: 5432 });
    } catch (error) {
      console.error('Failed to connect to database:', error.message);
      throw error;
    }
  });

  afterAll(async () => {
    if (connection) {
      await connection.close();
    }
  });

  it('should have an active connection', () => {
    expect(connection).toBeDefined();
    expect(connection.isConnected).toBe(true);
  });

  it('should query the health table', async () => {
    const result = await connection.query('SELECT 1 AS healthy');
    expect(result.rows).toHaveLength(1);
    expect(result.rows[0].healthy).toBe(1);
  });
});
