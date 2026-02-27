describe('auth', () => {
  test('should login', () => {
    expect(login('user', 'pass')).toBe(true);
  });
});
