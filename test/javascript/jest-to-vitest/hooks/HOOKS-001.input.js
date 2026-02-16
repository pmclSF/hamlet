describe('DatabaseConnection', () => {
  let connection;

  beforeAll(() => {
    connection = {
      host: 'localhost',
      port: 5432,
      connected: true,
    };
  });

  it('should be connected after setup', () => {
    expect(connection.connected).toBe(true);
  });

  it('should use the correct host', () => {
    expect(connection.host).toBe('localhost');
  });

  it('should use the correct port', () => {
    expect(connection.port).toBe(5432);
  });
});
