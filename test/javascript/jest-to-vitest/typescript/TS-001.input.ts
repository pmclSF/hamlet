interface User {
  id: number;
  name: string;
  email: string;
}

describe('UserService', () => {
  it('should create a user', () => {
    const user: User = { id: 1, name: 'Alice', email: 'alice@test.com' };
    expect(user.name).toBe('Alice');
  });

  it('should validate email', () => {
    const isValid: boolean = validateEmail('test@example.com');
    expect(isValid).toBe(true);
  });
});
