import { describe, it, expect } from 'vitest';

describe('Async parameterized tests', () => {
  it.each([
    [1, 'Alice'],
    [2, 'Bob'],
    [3, 'Charlie'],
  ])('fetches user %i and expects name %s', async (id, expectedName) => {
    const user = await fetchUser(id);
    expect(user.name).toBe(expectedName);
    expect(user.id).toBe(id);
  });

  it.each([
    ['users', 3],
    ['posts', 10],
    ['comments', 25],
  ])('fetches %s endpoint and expects %i items', async (endpoint, count) => {
    const results = await fetchCollection(endpoint);
    expect(results).toHaveLength(count);
  });
});
