import { describe, it, expect } from 'vitest';
import { getUserById } from '@/services/user';
import { formatName } from '@/utils/format';

describe('User service integration', () => {
  it('fetches and formats a user name', async () => {
    const user = await getUserById(1);
    const formatted = formatName(user.firstName, user.lastName);
    expect(formatted).toBe('Alice Smith');
  });

  it('returns null for unknown user', async () => {
    const user = await getUserById(9999);
    expect(user).toBeNull();
  });
});
