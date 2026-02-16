/**
 * @description Tests the auth flow
 * @ticket JIRA-1234
 * @author QA Team
 */
describe('Auth', () => {
  it('should authenticate with valid credentials', () => {
    const credentials = { username: 'admin', password: 'secret' };
    const token = authenticate(credentials);
    expect(token).toBeDefined();
    expect(typeof token).toBe('string');
  });

  it('should reject invalid credentials', () => {
    const credentials = { username: 'admin', password: 'wrong' };
    expect(() => authenticate(credentials)).toThrow('Invalid credentials');
  });
});
