describe('date comparison', () => {
  it('should compare equal dates with toEqual', () => {
    const date1 = new Date('2024-01-01');
    const date2 = new Date('2024-01-01');
    expect(date1).toEqual(date2);
  });

  it('should detect different dates', () => {
    const date1 = new Date('2024-01-01');
    const date2 = new Date('2024-12-31');
    expect(date1).not.toEqual(date2);
  });

  it('should compare timestamps numerically', () => {
    const date = new Date('2024-06-15T12:00:00Z');
    expect(date.getTime()).toBe(1718452800000);
  });

  it('should handle date in object comparison', () => {
    const event = {
      name: 'Launch',
      date: new Date('2024-03-01'),
    };
    expect(event).toEqual({
      name: 'Launch',
      date: new Date('2024-03-01'),
    });
  });
});
