import { describe, it, expect } from 'vitest';
import type { User } from './types';
import { createUser } from './factory';

describe('User', () => {
  it('creates a user with default values', () => {
    const user = createUser();
    expect(user.name).toBeDefined();
    expect(user.email).toContain('@');
  });

  it('creates a user with custom name', () => {
    const user = createUser({ name: 'Bob' });
    expect(user.name).toBe('Bob');
  });

  it('assigns a unique id to each user', () => {
    const user1 = createUser();
    const user2 = createUser();
    expect(user1.id).not.toBe(user2.id);
  });
});
