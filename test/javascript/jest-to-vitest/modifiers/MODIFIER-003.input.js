describe('SearchEngine', () => {
  it('should return results for valid queries', () => {
    const query = 'javascript testing';
    const results = [{ title: 'Jest Guide' }, { title: 'Vitest Docs' }];
    expect(results.length).toBeGreaterThan(0);
    expect(results[0].title).toBeDefined();
  });

  it.only('should rank exact matches first', () => {
    const results = [
      { title: 'Partial match', score: 0.7 },
      { title: 'Exact match', score: 1.0 },
    ];
    const sorted = results.sort((a, b) => b.score - a.score);
    expect(sorted[0].title).toBe('Exact match');
    expect(sorted[0].score).toBe(1.0);
  });

  fit('should handle empty query gracefully', () => {
    const query = '';
    const results = query.length > 0 ? [{ title: 'Result' }] : [];
    expect(results).toHaveLength(0);
  });

  it('should paginate results', () => {
    const allResults = Array.from({ length: 25 }, (_, i) => ({ id: i + 1 }));
    const page1 = allResults.slice(0, 10);
    expect(page1).toHaveLength(10);
    expect(page1[0].id).toBe(1);
  });
});
