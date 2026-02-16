import { describe, it, expect } from 'vitest';

interface User {
  id: number;
  name: string;
  email: string;
}

function createMock<T>(partial: Partial<T>): T {
  return partial as T;
}

describe('Generic helpers', () => {
  it('should create typed mock', () => {
    const user = createMock<User>({ name: 'Alice' });
    expect(user.name).toBe('Alice');
  });

  it('should create typed mock with multiple fields', () => {
    const user = createMock<User>({ id: 1, name: 'Bob' });
    expect(user.id).toBe(1);
    expect(user.name).toBe('Bob');
  });
});
