describe('test', () => {
  it('multiple', () => {
    expect(1).toBe(1);
    expect('hello').toHaveLength(5);
    expect([1, 2]).toContain(1);
    expect({ a: 1 }).toHaveProperty('a');
  });
});
