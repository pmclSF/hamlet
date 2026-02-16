describe('test', () => {
  it('close to', () => {
    expect(0.1 + 0.2).toBeCloseTo(0.3, 0.01);
  });
});
