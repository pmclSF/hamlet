describe('Array utilities', () => {
  it('should flatten nested arrays', () => {
    const nested = [[1, 2], [3, 4], [5]];
    const result = nested.flat();
    expect(result).toEqual([1, 2, 3, 4, 5]);
  });

  it('should remove duplicates', () => {
    const arr = [1, 2, 2, 3, 3, 3];
    const unique = [...new Set(arr)];
    expect(unique).toEqual([1, 2, 3]);
  });

  it('should find an element', () => {
    const items = [{ id: 1 }, { id: 2 }, { id: 3 }];
    const found = items.find(item => item.id === 2);
    expect(found).toEqual({ id: 2 });
  });
});
