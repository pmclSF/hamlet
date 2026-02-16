describe('ResourceManager', () => {
  let resource;

  beforeEach(() => {
    resource = { open: true, data: [] };
  });

  afterEach(() => {
    resource.open = false;
    resource.data = [];
  });

  it('should use the resource', () => {
    resource.data.push('item1');
    expect(resource.data).toHaveLength(1);
    expect(resource.open).toBe(true);
  });

  it('should handle failure gracefully', () => {
    resource.data.push('bad-item');
    expect(resource.open).toBe(true);
    throw new Error('Something went wrong');
  });
});
