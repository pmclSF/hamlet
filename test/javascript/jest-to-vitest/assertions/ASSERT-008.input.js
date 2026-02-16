describe('Pagination', () => {
  it('should return the correct page size', () => {
    const page = getPage(1, 10);
    expect(page.items).toHaveLength(10);
  });

  it('should return fewer items on the last page', () => {
    const page = getPage(3, 10);
    expect(page.items).toHaveLength(5);
  });

  it('should return an empty array when page is out of range', () => {
    const page = getPage(100, 10);
    expect(page.items).toHaveLength(0);
  });

  it('should handle string length', () => {
    const slug = generateSlug('Hello World');
    expect(slug).toHaveLength(11);
  });
});
