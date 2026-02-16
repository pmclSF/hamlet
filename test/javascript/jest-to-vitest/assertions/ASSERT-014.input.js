describe('DatabaseConnection', () => {
  it('should throw with a specific message for invalid host', () => {
    expect(() => connect({ host: '' })).toThrow('Host cannot be empty');
  });

  it('should throw matching a regex pattern for port errors', () => {
    expect(() => connect({ host: 'localhost', port: -1 })).toThrow(/invalid port/i);
  });

  it('should throw with the table name in the message', () => {
    expect(() => query('SELECT * FROM missing_table')).toThrow('missing_table');
  });

  it('should mention credentials in auth failures', () => {
    expect(() => authenticate('bad-user', 'bad-pass')).toThrow(/authentication failed/i);
  });
});
