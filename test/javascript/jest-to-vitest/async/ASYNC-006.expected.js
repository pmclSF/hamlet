import { describe, it, expect } from 'vitest';

describe('Parallel fetching', () => {
  it('fetches multiple resources concurrently', async () => {
    const [users, posts, comments] = await Promise.all([
      fetchUsers(),
      fetchPosts(),
      fetchComments(),
    ]);
    expect(users).toHaveLength(3);
    expect(posts).toHaveLength(10);
    expect(comments).toHaveLength(25);
  });

  it('handles partial failures with allSettled', async () => {
    const results = await Promise.allSettled([
      fetchUsers(),
      fetchInvalidResource(),
    ]);
    expect(results[0].status).toBe('fulfilled');
    expect(results[1].status).toBe('rejected');
  });
});
