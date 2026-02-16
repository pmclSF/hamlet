describe('assertions', () => {
  it('basic assertions', () => {
    expect(1).toBe(1);
    expect({ a: 1 }).toEqual({ a: 1 });
    expect(true).toBeTruthy();
    expect(null).toBeNull();
    expect('hello').toBeDefined();
  });
});
