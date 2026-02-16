function createUser(overrides = {}) {
  return { id: 1, name: 'Alice', email: 'alice@example.com', ...overrides };
}

function validateEmail(email) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

function validateUser(user) {
  return user.id > 0 && user.name.length > 0 && validateEmail(user.email);
}

describe('UserService', () => {
  it('should reject invalid email', () => {
    const user = createUser({ email: 'invalid' });
    expect(validateEmail(user.email)).toBe(false);
  });

  it('should accept valid user', () => {
    const user = createUser();
    expect(validateUser(user)).toBe(true);
  });

  it('should allow overriding name', () => {
    const user = createUser({ name: 'Bob' });
    expect(user.name).toBe('Bob');
    expect(user.email).toBe('alice@example.com');
  });

  it('should allow overriding multiple fields', () => {
    const user = createUser({ id: 99, name: 'Charlie', email: 'charlie@test.com' });
    expect(user.id).toBe(99);
    expect(validateUser(user)).toBe(true);
  });
});
