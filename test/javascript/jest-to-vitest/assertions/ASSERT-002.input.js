describe('UserService', () => {
  it('should return a user object with all fields', () => {
    const user = createUser('Alice', 30);
    expect(user).toEqual({ name: 'Alice', age: 30 });
  });

  it('should merge preferences with defaults', () => {
    const prefs = mergePreferences({ theme: 'dark' });
    expect(prefs).toEqual({
      theme: 'dark',
      language: 'en',
      notifications: true,
    });
  });

  it('should handle nested objects', () => {
    const config = buildConfig({ db: { host: 'localhost', port: 5432 } });
    expect(config.db).toEqual({ host: 'localhost', port: 5432 });
  });
});
