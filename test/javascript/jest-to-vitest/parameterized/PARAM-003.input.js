describe.each([
  ['mobile', 375],
  ['tablet', 768],
  ['desktop', 1024],
])('on %s (width: %i)', (viewport, width) => {
  it('should have a positive width', () => {
    expect(width).toBeGreaterThan(0);
  });

  it('should have a defined viewport name', () => {
    expect(viewport).toBeDefined();
    expect(typeof viewport).toBe('string');
  });

  it('should render the layout correctly', () => {
    const layout = getLayout(width);
    expect(layout).toBeTruthy();
  });
});
