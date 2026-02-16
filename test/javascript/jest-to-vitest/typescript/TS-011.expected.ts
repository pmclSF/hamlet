import { describe, it, expect } from 'vitest';

interface User {
  id: number;
  name: string;
  address?: {
    city: string;
    zip: string;
  };
}

describe('Strict null checks', () => {
  it('should use non-null assertion', () => {
    const users: User[] = [{ id: 1, name: 'Alice', address: { city: 'NYC', zip: '10001' } }];
    const first = users.find(u => u.id === 1);
    expect(first!.name).toBe('Alice');
  });

  it('should use optional chaining', () => {
    const user: User = { id: 1, name: 'Bob' };
    expect(user.address?.city).toBeUndefined();
  });

  it('should combine optional chaining with nullish coalescing', () => {
    const user: User = { id: 2, name: 'Carol' };
    const city: string = user.address?.city ?? 'Unknown';
    expect(city).toBe('Unknown');
  });
});
