describe('UserAPI', () => {
  it('fetches user data with promises', () => {
    return fetchUser(1).then(data => {
      expect(data.name).toBe('Alice');
      expect(data.email).toBeDefined();
    });
  });

  it('handles chained promises', () => {
    return fetchUser(1)
      .then(user => fetchPosts(user.id))
      .then(posts => {
        expect(posts).toHaveLength(5);
      });
  });
});
